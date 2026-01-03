package complete

import (
	"sort"
	"strings"

	"github.com/pierre-borckmans/scql/pkg/analyze"
	"github.com/pierre-borckmans/scql/pkg/schema"
)

// GetCompletions returns completion items for the given context.
// For group information, use GetCompletionsResult instead.
func GetCompletions(ctx *CompletionContext) []CompletionItem {
	result := GetCompletionsResult(ctx)
	return result.Items
}

// GetCompletionsResult returns completions with group definitions.
func GetCompletionsResult(ctx *CompletionContext) CompletionResult {
	return GetCompletionsResultWithOptions(ctx, DefaultOptions())
}

// GetCompletionsWithOptions returns completion items with custom options.
// For group information, use GetCompletionsResultWithOptions instead.
func GetCompletionsWithOptions(ctx *CompletionContext, opts *CompletionOptions) []CompletionItem {
	result := GetCompletionsResultWithOptions(ctx, opts)
	return result.Items
}

// GetCompletionsResultWithOptions returns completions with group definitions and custom options.
func GetCompletionsResultWithOptions(ctx *CompletionContext, opts *CompletionOptions) CompletionResult {
	if ctx == nil {
		return CompletionResult{}
	}

	// Create group registry to collect groups
	registry := NewGroupRegistry()

	// Detect the completion context
	detected := DetectContext(ctx.Query, ctx.Position)

	// Get completions based on context
	items := getCompletionsForContext(detected, ctx.Schema, ctx.DefaultKeyspace, opts, registry)

	// Apply ANTLR filter to remove grammatically invalid keywords
	if opts.UseANTLRFilter && len(items) > 0 {
		items = filterWithANTLR(ctx.Query, ctx.Position, items, detected.Prefix)
	}

	// Filter by prefix
	if detected.Prefix != "" {
		items = filterByPrefix(items, detected.Prefix)
	}

	// Sort by priority
	sortCompletions(items)

	// Apply limit
	if opts.MaxItems > 0 && len(items) > opts.MaxItems {
		items = items[:opts.MaxItems]
	}

	// Determine if multi-select is allowed based on context
	allowMultiSelect := detected.Type == ContextAfterSelect || detected.Type == ContextInColumnList

	// Extract selected columns for multi-select UI (only for SELECT context)
	var selectedColumns []string
	if allowMultiSelect && detected.Type == ContextAfterSelect {
		refs, _, _ := analyze.ExtractReferences(ctx.Query)
		if refs != nil {
			selectedColumns = refs.SelectColumns
		}
	}

	return CompletionResult{
		Groups:           registry.Groups(),
		Items:            items,
		Context:          detected.Type,
		AllowMultiSelect: allowMultiSelect,
		SelectedColumns:  selectedColumns,
	}
}

// getCompletionsForContext generates completions based on detected context.
func getCompletionsForContext(ctx *DetectedContext, s *schema.Schema, defaultKs string, opts *CompletionOptions, registry *GroupRegistry) []CompletionItem {
	var items []CompletionItem

	switch ctx.Type {
	case ContextStatementStart:
		if opts.IncludeSnippets {
			items = append(items, Snippets...)
		}
		if opts.IncludeKeywords {
			items = append(items, StatementKeywords...)
		}

	case ContextAfterSelect:
		// Columns, *, functions, and FROM (to finish column list)
		items = append(items, CompletionItem{Label: "*", Kind: KindKeyword, Detail: "All columns", SortPriority: 0})
		items = append(items, CompletionItem{Label: "FROM", Kind: KindKeyword, Detail: "Specify table", SortPriority: 100})
		if opts.IncludeFunctions {
			items = append(items, getFunctionCompletions(registry)...)
		}
		// Add table columns if we have schema context
		if s != nil && ctx.Table != "" {
			items = append(items, getColumnCompletions(s, ctx.Keyspace, ctx.Table, defaultKs, registry)...)
		}

	case ContextAfterSelectColumns:
		// After SELECT columns - suggest FROM
		items = append(items, CompletionItem{Label: "FROM", Kind: KindKeyword, Detail: "Specify table", SortPriority: 1})

	case ContextAfterFrom, ContextAfterInsertInto, ContextAfterUpdate:
		// Tables and keyspaces
		if s != nil {
			items = append(items, getTableCompletions(s, ctx.Keyspace, defaultKs, registry)...)
			items = append(items, getKeyspaceCompletions(s, registry)...)
		}

	case ContextAfterDot:
		// After "keyspace." - suggest tables
		// After "table." - suggest columns (less common in CQL)
		if s != nil && ctx.Keyspace != "" {
			// Try as keyspace - get tables
			items = append(items, getTableCompletions(s, ctx.Keyspace, "", registry)...)
		}

	case ContextAfterSelectTable:
		// After SELECT ... FROM table - suggest valid SELECT clauses
		if opts.IncludeKeywords {
			items = append(items, SelectTableKeywords...)
		}

	case ContextAfterWhere:
		// After WHERE keyword - suggest columns for conditions
		if s != nil && ctx.Table != "" {
			items = append(items, getColumnCompletions(s, ctx.Keyspace, ctx.Table, defaultKs, registry)...)
		}
		if opts.IncludeFunctions {
			// token() is common in WHERE
			item := CompletionItem{Label: "token()", Kind: KindFunction, Detail: "Partition token", InsertText: "token()", SortPriority: 50}
			item.Groups = []string{registry.RegisterFunctionCategory("token")}
			items = append(items, item)
		}

	case ContextAfterAnd:
		// After a complete WHERE condition - suggest AND or next clauses
		if s != nil && ctx.Table != "" {
			items = append(items, getColumnCompletions(s, ctx.Keyspace, ctx.Table, defaultKs, registry)...)
		}
		if opts.IncludeKeywords {
			items = append(items, WhereClauseKeywords...)
		}
		if opts.IncludeFunctions {
			item := CompletionItem{Label: "token()", Kind: KindFunction, Detail: "Partition token", InsertText: "token()", SortPriority: 50}
			item.Groups = []string{registry.RegisterFunctionCategory("token")}
			items = append(items, item)
		}

	case ContextAfterColumn:
		// After column name in WHERE - suggest operators
		items = append(items, Operators...)

	case ContextAfterUpdateSet:
		// After UPDATE table SET col = value - suggest WHERE, USING, IF
		if opts.IncludeKeywords {
			items = append(items, UpdateSetKeywords...)
		}

	case ContextAfterDeleteFrom:
		// After DELETE FROM table - suggest WHERE, USING, IF
		if opts.IncludeKeywords {
			items = append(items, DeleteFromKeywords...)
		}

	case ContextAfterOperator:
		// Bind marker is always useful
		items = append(items, CompletionItem{Label: "?", Kind: KindSnippet, Detail: "Bind marker", SortPriority: 1})

		// Get column type if available
		columnType := ""
		if s != nil && ctx.Table != "" && ctx.Column != "" {
			columnType = getColumnType(s, ctx.Keyspace, ctx.Table, defaultKs, ctx.Column)
		}

		// Add type-aware completions
		items = append(items, getValueCompletionsForType(columnType, opts.IncludeFunctions, registry)...)

	case ContextAfterSet:
		// Columns for UPDATE SET
		if s != nil && ctx.Table != "" {
			items = append(items, getColumnCompletions(s, ctx.Keyspace, ctx.Table, defaultKs, registry)...)
		}

	case ContextAfterDelete:
		// Can be column list or FROM
		items = append(items, CompletionItem{Label: "FROM", Kind: KindKeyword, Detail: "Specify table", SortPriority: 1})
		if s != nil && ctx.Table != "" {
			items = append(items, getColumnCompletions(s, ctx.Keyspace, ctx.Table, defaultKs, registry)...)
		}

	case ContextAfterOrderBy, ContextAfterGroupBy:
		// Columns - ASC/DESC are only valid AFTER a column name, not directly after ORDER BY
		if s != nil && ctx.Table != "" {
			items = append(items, getColumnCompletions(s, ctx.Keyspace, ctx.Table, defaultKs, registry)...)
		}

	case ContextAfterCreate:
		if opts.IncludeKeywords {
			items = append(items, CreateKeywords...)
		}

	case ContextAfterAlter:
		if opts.IncludeKeywords {
			items = append(items, AlterKeywords...)
		}

	case ContextAfterDrop:
		if opts.IncludeKeywords {
			items = append(items, DropKeywords...)
		}

	case ContextAfterUse:
		// Keyspaces
		if s != nil {
			items = append(items, getKeyspaceCompletions(s, registry)...)
		}

	case ContextAfterDescribe:
		// DESCRIBE can be followed by object type or specific object name
		if opts.IncludeKeywords {
			items = append(items, DescribeKeywords...)
		}
		// Also suggest keyspaces and tables directly
		if s != nil {
			items = append(items, getKeyspaceCompletions(s, registry)...)
			items = append(items, getTableCompletions(s, ctx.Keyspace, defaultKs, registry)...)
		}

	case ContextAfterPrune:
		// PRUNE can only be followed by MATERIALIZED VIEW
		if opts.IncludeKeywords {
			items = append(items, PruneKeywords...)
		}

	case ContextInColumnList:
		// Columns for INSERT column list
		if s != nil && ctx.Table != "" {
			items = append(items, getColumnCompletions(s, ctx.Keyspace, ctx.Table, defaultKs, registry)...)
		}

	case ContextAfterValues:
		// Values - suggest placeholders and functions
		items = append(items, CompletionItem{Label: "?", Kind: KindSnippet, Detail: "Bind marker", SortPriority: 1})
		if opts.IncludeFunctions {
			// Common value functions with groups
			uuidItem := CompletionItem{Label: "uuid()", Kind: KindFunction, Detail: "Generate UUID", InsertText: "uuid()", SortPriority: 10}
			uuidItem.Groups = []string{registry.RegisterFunctionCategory("uuid")}
			items = append(items, uuidItem)

			nowItem := CompletionItem{Label: "now()", Kind: KindFunction, Detail: "Current timeuuid", InsertText: "now()", SortPriority: 11}
			nowItem.Groups = []string{registry.RegisterFunctionCategory("now")}
			items = append(items, nowItem)
		}

	case ContextInTypeSpec:
		if opts.IncludeTypes {
			items = append(items, getTypeCompletions(registry)...)
		}

	case ContextAfterLimit:
		// After LIMIT, only a number is expected - no completions

	case ContextAfterLimitValue:
		// After LIMIT <number>, only ALLOW FILTERING is valid
		if opts.IncludeKeywords {
			items = append(items, CompletionItem{Label: "ALLOW FILTERING", Kind: KindKeyword, Detail: "Allow full table scan", SortPriority: 1})
		}

	case ContextUnknown:
		// Provide general clause keywords
		if opts.IncludeKeywords {
			items = append(items, ClauseKeywords...)
		}
	}

	return items
}

// getKeyspaceCompletions returns completions for keyspaces from schema.
func getKeyspaceCompletions(s *schema.Schema, registry *GroupRegistry) []CompletionItem {
	if s == nil {
		return nil
	}

	// Register the keyspaces category group
	groupID := registry.RegisterSchemaCategory(GroupCatKeyspaces)

	var items []CompletionItem
	for _, name := range s.KeyspaceNames() {
		items = append(items, CompletionItem{
			Label:        name,
			Kind:         KindKeyspace,
			Detail:       "Keyspace",
			SortPriority: 10,
			Groups:       []string{groupID},
		})
	}
	return items
}

// getTableCompletions returns completions for tables and materialized views from schema.
// MVs are queryable just like tables in CQL.
// If keyspace is specified, only returns tables/views from that keyspace.
// Otherwise, returns tables/views from all keyspaces:
//   - Items in defaultKs are shown without prefix (higher priority)
//   - Items in other keyspaces are shown with keyspace.name prefix
func getTableCompletions(s *schema.Schema, keyspace, defaultKs string, registry *GroupRegistry) []CompletionItem {
	if s == nil {
		return nil
	}

	var items []CompletionItem
	ksPriority := 0

	// If explicit keyspace is specified (e.g., after "myks."), only show tables/views from that keyspace
	if keyspace != "" {
		ks := s.GetKeyspace(keyspace)
		if ks != nil {
			ksGroupID := registry.RegisterKeyspace(keyspace, ksPriority)

			// Tables
			for _, name := range ks.TableNames() {
				tbl := ks.GetTable(name)
				detail := "Table"
				if tbl != nil && len(tbl.PartitionKey) > 0 {
					detail = "Table (PK: " + strings.Join(tbl.PartitionKey, ", ") + ")"
				}
				items = append(items, CompletionItem{
					Label:        name,
					Kind:         KindTable,
					Detail:       detail,
					SortPriority: 5,
					Groups:       []string{ksGroupID},
				})
			}
			// Materialized views
			for _, name := range ks.MaterializedViewNames() {
				mv := ks.GetMaterializedView(name)
				detail := "Materialized View"
				if mv != nil && mv.BaseTable != "" {
					detail = "View on " + mv.BaseTable
				}
				items = append(items, CompletionItem{
					Label:        name,
					Kind:         KindView,
					Detail:       detail,
					SortPriority: 6, // Slightly lower than tables
					Groups:       []string{ksGroupID},
				})
			}
		}
		return items
	}

	// No explicit keyspace - show tables/views from ALL keyspaces
	// Items in default keyspace get higher priority and no prefix
	// Items in other keyspaces get keyspace.name format
	for _, ksName := range s.KeyspaceNames() {
		ks := s.GetKeyspace(ksName)
		if ks == nil {
			continue
		}

		isDefault := ksName == defaultKs

		// Register keyspace group (default keyspace gets priority 0)
		priority := ksPriority
		if !isDefault {
			priority = ksPriority + 10
		}
		ksGroupID := registry.RegisterKeyspace(ksName, priority)
		ksPriority++

		// Tables
		for _, name := range ks.TableNames() {
			tbl := ks.GetTable(name)

			var label, detail string
			var sortPriority int

			if isDefault {
				label = name
				detail = "Table"
				if tbl != nil && len(tbl.PartitionKey) > 0 {
					detail = "Table (PK: " + strings.Join(tbl.PartitionKey, ", ") + ")"
				}
				sortPriority = 5
			} else {
				label = ksName + "." + name
				detail = "Table in " + ksName
				sortPriority = 10
			}

			items = append(items, CompletionItem{
				Label:        label,
				Kind:         KindTable,
				Detail:       detail,
				SortPriority: sortPriority,
				FilterText:   name,
				Groups:       []string{ksGroupID},
			})
		}

		// Materialized views
		for _, name := range ks.MaterializedViewNames() {
			mv := ks.GetMaterializedView(name)

			var label, detail string
			var sortPriority int

			if isDefault {
				label = name
				detail = "Materialized View"
				if mv != nil && mv.BaseTable != "" {
					detail = "View on " + mv.BaseTable
				}
				sortPriority = 6 // Slightly lower than tables
			} else {
				label = ksName + "." + name
				detail = "View in " + ksName
				sortPriority = 11
			}

			items = append(items, CompletionItem{
				Label:        label,
				Kind:         KindView,
				Detail:       detail,
				SortPriority: sortPriority,
				FilterText:   name,
				Groups:       []string{ksGroupID},
			})
		}
	}

	return items
}

// getFunctionCompletions returns function completions with category groups.
func getFunctionCompletions(registry *GroupRegistry) []CompletionItem {
	items := make([]CompletionItem, 0, len(CQLFunctions))
	for _, fn := range CQLFunctions {
		item := fn // Copy
		catID := registry.RegisterFunctionCategory(strings.ToLower(strings.TrimSuffix(fn.Label, "()")))
		item.Groups = []string{catID}
		items = append(items, item)
	}
	return items
}

// getTypeCompletions returns type completions with category groups.
func getTypeCompletions(registry *GroupRegistry) []CompletionItem {
	items := make([]CompletionItem, 0, len(CQLTypes))
	for _, typ := range CQLTypes {
		item := typ // Copy
		catID := registry.RegisterTypeCategory(strings.ToLower(typ.Label))
		item.Groups = []string{catID}
		items = append(items, item)
	}
	return items
}

// getColumnCompletions returns completions for columns from schema.
func getColumnCompletions(s *schema.Schema, keyspace, table, defaultKs string, registry *GroupRegistry) []CompletionItem {
	if s == nil || table == "" {
		return nil
	}

	// Determine keyspace
	ksName := keyspace
	if ksName == "" {
		ksName = defaultKs
	}
	if ksName == "" {
		return nil
	}

	ks := s.GetKeyspace(ksName)
	if ks == nil {
		return nil
	}

	tbl := ks.GetTable(table)
	if tbl == nil {
		return nil
	}

	// Register columns category group and source table group
	colGroupID := registry.RegisterSchemaCategory(GroupCatColumns)
	srcGroupID := registry.RegisterSource(table, 0)

	var items []CompletionItem
	for _, col := range tbl.AllColumns() {
		detail := col.Type
		sortPriority := 20
		groups := []string{colGroupID, srcGroupID}

		// Highlight key columns and add key type groups
		if col.IsPartitionKey {
			detail = col.Type + " (partition key)"
			sortPriority = 1
			groups = append(groups, registry.RegisterKeyType("partition"))
		} else if col.IsClusteringKey {
			detail = col.Type + " (clustering key)"
			sortPriority = 2
			groups = append(groups, registry.RegisterKeyType("clustering"))
		} else if col.IsStatic {
			detail = col.Type + " (static)"
			sortPriority = 10
			groups = append(groups, registry.RegisterKeyType("static"))
		} else {
			groups = append(groups, registry.RegisterKeyType("regular"))
		}

		items = append(items, CompletionItem{
			Label:        col.Name,
			Kind:         KindColumn,
			Detail:       detail,
			SortPriority: sortPriority,
			Groups:       groups,
		})
	}

	return items
}

// filterWithANTLR filters keyword completions using ANTLR validation.
// Only specific keywords are filtered - snippets, columns, tables, etc. are kept as-is.
func filterWithANTLR(query string, position int, items []CompletionItem, prefix string) []CompletionItem {
	// Calculate the position without the prefix (where the completion will be inserted)
	insertPosition := position - len(prefix)
	if insertPosition < 0 {
		insertPosition = 0
	}

	// Truncate query to the insert position
	truncatedQuery := query
	if insertPosition < len(query) {
		truncatedQuery = query[:insertPosition]
	}

	result := make([]CompletionItem, 0, len(items))
	for _, item := range items {
		// Only filter keyword completions - not snippets (like "?") or schema items
		if item.Kind == KindKeyword {
			// Get the text that would be inserted
			checkText := item.Label
			// For multi-word keywords like "ORDER BY", just check the first word
			if spaceIdx := strings.Index(checkText, " "); spaceIdx > 0 {
				checkText = checkText[:spaceIdx]
			}

			// Validate using ANTLR
			if IsTokenValidAtPosition(truncatedQuery, len(truncatedQuery), checkText) {
				result = append(result, item)
			}
		} else {
			// Keep non-keyword items (snippets, columns, tables, functions, etc.)
			result = append(result, item)
		}
	}

	return result
}

// filterByPrefix filters completions to those matching the prefix.
func filterByPrefix(items []CompletionItem, prefix string) []CompletionItem {
	if prefix == "" {
		return items
	}

	prefix = strings.ToLower(prefix)
	var filtered []CompletionItem

	for _, item := range items {
		filterText := item.FilterText
		if filterText == "" {
			filterText = item.Label
		}
		if strings.HasPrefix(strings.ToLower(filterText), prefix) {
			filtered = append(filtered, item)
		}
	}

	return filtered
}

// sortCompletions sorts completions by priority, then alphabetically.
func sortCompletions(items []CompletionItem) {
	sort.Slice(items, func(i, j int) bool {
		if items[i].SortPriority != items[j].SortPriority {
			return items[i].SortPriority < items[j].SortPriority
		}
		return strings.ToLower(items[i].Label) < strings.ToLower(items[j].Label)
	})
}

// getColumnType returns the type of a column from the schema.
func getColumnType(s *schema.Schema, keyspace, table, defaultKs, column string) string {
	if s == nil || table == "" || column == "" {
		return ""
	}

	ksName := keyspace
	if ksName == "" {
		ksName = defaultKs
	}
	if ksName == "" {
		return ""
	}

	ks := s.GetKeyspace(ksName)
	if ks == nil {
		return ""
	}

	tbl := ks.GetTable(table)
	if tbl == nil {
		return ""
	}

	for _, col := range tbl.AllColumns() {
		if strings.EqualFold(col.Name, column) {
			return strings.ToLower(col.Type)
		}
	}

	return ""
}

// getValueCompletionsForType returns value completions appropriate for the given type.
func getValueCompletionsForType(columnType string, includeFunctions bool, registry *GroupRegistry) []CompletionItem {
	var items []CompletionItem

	// Normalize the type (handle generics like map<text,text>, list<int>, etc.)
	baseType := columnType
	if idx := strings.Index(columnType, "<"); idx != -1 {
		baseType = columnType[:idx]
	}

	// null is valid for all types
	items = append(items, CompletionItem{Label: "null", Kind: KindKeyword, Detail: "Null value", SortPriority: 100})

	// Helper to create function item with group
	addFn := func(label, detail, insertText string, priority int, funcName string) {
		item := CompletionItem{Label: label, Kind: KindFunction, Detail: detail, InsertText: insertText, SortPriority: priority}
		if registry != nil {
			item.Groups = []string{registry.RegisterFunctionCategory(funcName)}
		}
		items = append(items, item)
	}

	switch baseType {
	case "boolean":
		items = append(items, CompletionItem{Label: "true", Kind: KindKeyword, Detail: "Boolean true", SortPriority: 2})
		items = append(items, CompletionItem{Label: "false", Kind: KindKeyword, Detail: "Boolean false", SortPriority: 3})

	case "int", "bigint", "smallint", "tinyint", "varint", "counter":
		// Numeric - no special completions, user types the number

	case "float", "double", "decimal":
		// Numeric decimal - no special completions

	case "text", "varchar", "ascii":
		// String - suggest empty string
		items = append(items, CompletionItem{Label: "''", Kind: KindSnippet, Detail: "Empty string", SortPriority: 10})

	case "timestamp":
		if includeFunctions {
			addFn("toTimestamp(now())", "Current timestamp", "toTimestamp(now())", 2, "totimestamp")
			addFn("currentTimestamp()", "Current timestamp", "currentTimestamp()", 3, "currenttimestamp")
		}

	case "date":
		if includeFunctions {
			addFn("currentDate()", "Current date", "currentDate()", 2, "currentdate")
			addFn("toDate(now())", "Current date", "toDate(now())", 3, "todate")
		}

	case "time":
		if includeFunctions {
			addFn("currentTime()", "Current time", "currentTime()", 2, "currenttime")
		}

	case "timeuuid":
		if includeFunctions {
			addFn("now()", "Current timeuuid", "now()", 2, "now")
			addFn("currentTimeUUID()", "Current timeuuid", "currentTimeUUID()", 3, "currenttimeuuid")
			addFn("minTimeuuid()", "Min timeuuid for timestamp", "minTimeuuid()", 10, "mintimeuuid")
			addFn("maxTimeuuid()", "Max timeuuid for timestamp", "maxTimeuuid()", 11, "maxtimeuuid")
		}

	case "uuid":
		if includeFunctions {
			addFn("uuid()", "Generate random UUID", "uuid()", 2, "uuid")
		}

	case "blob":
		if includeFunctions {
			addFn("textAsBlob()", "Convert text to blob", "textAsBlob()", 10, "textasblob")
		}

	case "inet":
		// IP address - no special completions

	case "duration":
		// Duration literal examples
		items = append(items, CompletionItem{Label: "1h", Kind: KindSnippet, Detail: "1 hour duration", SortPriority: 10})
		items = append(items, CompletionItem{Label: "30m", Kind: KindSnippet, Detail: "30 minutes duration", SortPriority: 11})
		items = append(items, CompletionItem{Label: "1d", Kind: KindSnippet, Detail: "1 day duration", SortPriority: 12})

	case "list", "set":
		items = append(items, CompletionItem{Label: "[]", Kind: KindSnippet, Detail: "Empty collection", InsertText: "[]", SortPriority: 10})

	case "map":
		items = append(items, CompletionItem{Label: "{}", Kind: KindSnippet, Detail: "Empty map", InsertText: "{}", SortPriority: 10})

	case "tuple":
		items = append(items, CompletionItem{Label: "()", Kind: KindSnippet, Detail: "Tuple", InsertText: "()", SortPriority: 10})

	case "":
		// Unknown type - provide generic completions
		items = append(items, CompletionItem{Label: "true", Kind: KindKeyword, Detail: "Boolean true", SortPriority: 10})
		items = append(items, CompletionItem{Label: "false", Kind: KindKeyword, Detail: "Boolean false", SortPriority: 11})
		if includeFunctions {
			addFn("uuid()", "Generate UUID", "uuid()", 20, "uuid")
			addFn("now()", "Current timeuuid", "now()", 21, "now")
			addFn("toTimestamp(now())", "Current timestamp", "toTimestamp(now())", 22, "totimestamp")
		}
	}

	return items
}
