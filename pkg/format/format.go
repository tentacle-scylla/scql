package format

import (
	"strings"

	"github.com/antlr4-go/antlr/v4"

	parser "github.com/tentacle-scylla/scql/gen/parser"
	"github.com/tentacle-scylla/scql/pkg/parse"
	"github.com/tentacle-scylla/scql/pkg/types"
)

// Style defines the formatting style for CQL output
type Style int

const (
	// Compact produces single-line, minimal whitespace output
	Compact Style = iota

	// Pretty produces multi-line, indented output
	Pretty
)

// Options configures the formatter behavior
type Options struct {
	Style             Style
	IndentString      string // Default: "    " (4 spaces)
	UppercaseKeywords bool   // Default: false
}

// DefaultOptions returns sensible defaults for ScyllaDB-style formatting
func DefaultOptions() Options {
	return Options{
		Style:             Pretty,
		IndentString:      "    ",
		UppercaseKeywords: false, // ScyllaDB uses lowercase types
	}
}

// CompactOptions returns options for compact formatting
func CompactOptions() Options {
	return Options{
		Style:             Compact,
		IndentString:      "",
		UppercaseKeywords: true,
	}
}

// Format formats a parsed CQL statement according to the given options
func Format(result *parse.Result, opts Options) string {
	if result == nil || result.Tree == nil || result.Tokens == nil {
		return ""
	}

	f := &formatter{
		opts:   opts,
		tokens: result.Tokens,
		result: result,
	}

	return f.format()
}

// String parses and formats a CQL string
func String(input string, opts Options) (string, error) {
	result := parse.Parse(input)
	if result.HasErrors() {
		return "", result.Errors
	}
	return Format(result, opts), nil
}

// PrettyString is a convenience function for pretty formatting
func PrettyString(input string) (string, error) {
	return String(input, DefaultOptions())
}

// CompactString is a convenience function for compact formatting
func CompactString(input string) (string, error) {
	return String(input, CompactOptions())
}

// formatter handles the formatting logic
type formatter struct {
	opts   Options
	tokens *antlr.CommonTokenStream
	result *parse.Result
	indent int
	output strings.Builder
}

func (f *formatter) format() string {
	f.output.Reset()
	f.indent = 0

	// Use AST-aware formatting for specific statement types
	if f.result.Cql != nil && f.opts.Style == Pretty {
		switch f.result.Type {
		// DDL - Create statements
		case types.StatementCreateTable:
			return f.formatCreateTable(f.result.Cql.CreateTable())
		case types.StatementCreateType:
			return f.formatCreateType(f.result.Cql.CreateType())
		case types.StatementCreateKeyspace:
			return f.formatCreateKeyspace(f.result.Cql.CreateKeyspace())
		case types.StatementAlterKeyspace:
			return f.formatAlterKeyspace(f.result.Cql.AlterKeyspace())
		case types.StatementCreateMaterializedView:
			return f.formatCreateMaterializedView(f.result.Cql.CreateMaterializedView())
		case types.StatementCreateIndex:
			return f.formatCreateIndex(f.result.Cql.CreateIndex())

		// DML statements
		case types.StatementSelect:
			return f.formatSelect(f.result.Cql.Select_())
		case types.StatementInsert:
			// Check if this is actually a batch (INSERT with BeginBatch)
			if insert := f.result.Cql.Insert(); insert != nil && insert.BeginBatch() != nil {
				return f.formatTokenBased() // Use token-based for batch
			}
			return f.formatInsert(f.result.Cql.Insert())
		case types.StatementUpdate:
			// Check if this is actually a batch (UPDATE with BeginBatch)
			if update := f.result.Cql.Update(); update != nil && update.BeginBatch() != nil {
				return f.formatTokenBased() // Use token-based for batch
			}
			return f.formatUpdate(f.result.Cql.Update())
		case types.StatementDelete:
			// Check if this is actually a batch (DELETE with BeginBatch)
			if del := f.result.Cql.Delete_(); del != nil && del.BeginBatch() != nil {
				return f.formatTokenBased() // Use token-based for batch
			}
			return f.formatDelete(f.result.Cql.Delete_())
		case types.StatementBatch:
			return f.formatBatch(f.result.Cql.Batch())
		}
	}

	// Fall back to token-based formatting for other statements
	return f.formatTokenBased()
}

// formatCreateTable formats a CREATE TABLE statement with proper structure
func (f *formatter) formatCreateTable(ctx parser.ICreateTableContext) string {
	if ctx == nil {
		return f.formatTokenBased()
	}

	var sb strings.Builder
	indent := f.opts.IndentString

	// CREATE TABLE [IF NOT EXISTS]
	sb.WriteString("CREATE TABLE")
	if ctx.IfNotExist() != nil {
		sb.WriteString(" IF NOT EXISTS")
	}

	// [keyspace.]table
	sb.WriteString(" ")
	if ctx.Keyspace() != nil {
		sb.WriteString(f.formatNode(ctx.Keyspace()))
		sb.WriteString(".")
	}
	sb.WriteString(f.formatNode(ctx.Table()))

	// Column definitions
	sb.WriteString(" (\n")

	colDefList := ctx.ColumnDefinitionList()
	if colDefList != nil {
		allColDefs := colDefList.AllColumnDefinition()
		for i, colDef := range allColDefs {
			sb.WriteString(indent)
			sb.WriteString(f.formatColumnDefinition(colDef))
			// Add comma if not last or if PRIMARY KEY follows
			if i < len(allColDefs)-1 || colDefList.PrimaryKeyElement() != nil {
				sb.WriteString(",")
			}
			sb.WriteString("\n")
		}

		// PRIMARY KEY clause (if separate from column)
		if pkElem := colDefList.PrimaryKeyElement(); pkElem != nil {
			sb.WriteString(indent)
			sb.WriteString(f.formatPrimaryKeyElement(pkElem))
			sb.WriteString("\n")
		}
	}

	sb.WriteString(")")

	// WITH clause (on same line as closing paren, ScyllaDB style)
	if withElem := ctx.WithElement(); withElem != nil {
		sb.WriteString(" ")
		sb.WriteString(f.formatWithElement(withElem))
	}

	sb.WriteString(";")
	return sb.String()
}

// formatColumnDefinition formats a single column definition
func (f *formatter) formatColumnDefinition(ctx parser.IColumnDefinitionContext) string {
	if ctx == nil {
		return ""
	}

	var parts []string

	// Column name - preserve original case (don't uppercase even if it's a keyword)
	if col := ctx.Column(); col != nil {
		parts = append(parts, f.getOriginalTextRaw(col))
	}

	// Data type
	if dt := ctx.DataType(); dt != nil {
		parts = append(parts, f.formatDataType(dt))
	}

	// STATIC modifier
	if ctx.StaticColumn() != nil {
		parts = append(parts, "STATIC")
	}

	// PRIMARY KEY (inline)
	if ctx.PrimaryKeyColumn() != nil {
		parts = append(parts, "PRIMARY KEY")
	}

	return strings.Join(parts, " ")
}

// formatDataType formats a data type, uppercasing type names
func (f *formatter) formatDataType(ctx parser.IDataTypeContext) string {
	if ctx == nil {
		return ""
	}
	text := f.getOriginalText(ctx)
	if f.opts.UppercaseKeywords {
		// Uppercase common type names but preserve identifiers
		text = uppercaseDataType(text)
	}
	return text
}

// formatPrimaryKeyElement formats the PRIMARY KEY clause
func (f *formatter) formatPrimaryKeyElement(ctx parser.IPrimaryKeyElementContext) string {
	if ctx == nil {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("PRIMARY KEY ")

	// Get the primary key definition which contains partition and clustering keys
	if pkDef := ctx.PrimaryKeyDefinition(); pkDef != nil {
		sb.WriteString(f.formatPrimaryKeyDefinition(pkDef))
	}

	return sb.String()
}

// formatPrimaryKeyDefinition formats the key structure
func (f *formatter) formatPrimaryKeyDefinition(ctx parser.IPrimaryKeyDefinitionContext) string {
	if ctx == nil {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("(")

	// Check if we have compound partition key or simple
	if compoundKey := ctx.CompoundKey(); compoundKey != nil {
		// ((partition_cols), clustering_cols)
		sb.WriteString(f.formatNode(compoundKey))
	} else if compositeKey := ctx.CompositeKey(); compositeKey != nil {
		// (partition_col, clustering_cols)
		sb.WriteString(f.formatNode(compositeKey))
	} else {
		// Simple single column primary key
		sb.WriteString(f.getOriginalText(ctx))
		// Remove outer parens since we added them
		text := sb.String()
		if strings.HasPrefix(text, "((") {
			return text[1:]
		}
	}

	sb.WriteString(")")
	return sb.String()
}

// formatWithElement formats a WITH clause
func (f *formatter) formatWithElement(ctx parser.IWithElementContext) string {
	if ctx == nil {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("WITH ")

	// Collect all table options recursively
	parts := f.collectTableOptions(ctx.TableOptions())

	for i, part := range parts {
		if i > 0 {
			sb.WriteString("\n")
			sb.WriteString(f.opts.IndentString)
			sb.WriteString("AND ")
		}
		sb.WriteString(part)
	}

	return sb.String()
}

// collectTableOptions recursively collects all options from TableOptions
func (f *formatter) collectTableOptions(tableOpts parser.ITableOptionsContext) []string {
	if tableOpts == nil {
		return nil
	}

	var parts []string

	// Handle CLUSTERING ORDER BY first
	if clusteringOrder := tableOpts.ClusteringOrder(); clusteringOrder != nil {
		parts = append(parts, f.formatNode(clusteringOrder))
	}

	// Handle options at this level
	for _, opt := range tableOpts.AllTableOptionItem() {
		parts = append(parts, f.formatTableOptionItem(opt))
	}

	// Recursively collect from nested TableOptions
	if nested := tableOpts.TableOptions(); nested != nil {
		parts = append(parts, f.collectTableOptions(nested)...)
	}

	return parts
}

// formatTableOptionItem formats a single table option
func (f *formatter) formatTableOptionItem(ctx parser.ITableOptionItemContext) string {
	if ctx == nil {
		return ""
	}

	text := f.getOriginalText(ctx)

	// Uppercase the option name (before =)
	if idx := strings.Index(text, "="); idx != -1 {
		optName := strings.TrimSpace(text[:idx])
		optValue := strings.TrimSpace(text[idx+1:])
		return strings.ToLower(optName) + " = " + optValue
	}

	return text
}

// formatCreateType formats a CREATE TYPE statement
func (f *formatter) formatCreateType(ctx parser.ICreateTypeContext) string {
	if ctx == nil {
		return f.formatTokenBased()
	}

	var sb strings.Builder
	indent := f.opts.IndentString

	// CREATE TYPE [IF NOT EXISTS]
	sb.WriteString("CREATE TYPE")
	if ctx.IfNotExist() != nil {
		sb.WriteString(" IF NOT EXISTS")
	}

	// [keyspace.]type_name
	sb.WriteString(" ")
	if ctx.Keyspace() != nil {
		sb.WriteString(f.formatNode(ctx.Keyspace()))
		sb.WriteString(".")
	}
	sb.WriteString(f.formatNode(ctx.Type_()))

	// Field definitions
	sb.WriteString(" (\n")

	if fieldList := ctx.TypeMemberColumnList(); fieldList != nil {
		allFields := fieldList.AllColumn()
		allTypes := fieldList.AllDataType()

		for i := 0; i < len(allFields) && i < len(allTypes); i++ {
			sb.WriteString(indent)
			// Field name - preserve original case
			sb.WriteString(f.getOriginalTextRaw(allFields[i]))
			sb.WriteString(" ")
			sb.WriteString(f.formatDataType(allTypes[i]))
			if i < len(allFields)-1 {
				sb.WriteString(",")
			}
			sb.WriteString("\n")
		}
	}

	sb.WriteString(");")
	return sb.String()
}

// formatCreateKeyspace formats a CREATE KEYSPACE statement
func (f *formatter) formatCreateKeyspace(ctx parser.ICreateKeyspaceContext) string {
	if ctx == nil {
		return f.formatTokenBased()
	}

	var sb strings.Builder
	indent := f.opts.IndentString

	// CREATE KEYSPACE [IF NOT EXISTS]
	sb.WriteString("CREATE KEYSPACE")
	if ctx.IfNotExist() != nil {
		sb.WriteString(" IF NOT EXISTS")
	}

	// keyspace name
	sb.WriteString(" ")
	sb.WriteString(f.formatNode(ctx.Keyspace()))

	// WITH replication = {...}
	sb.WriteString("\n")
	sb.WriteString("WITH replication = ")
	if ctx.ReplicationList() != nil {
		sb.WriteString(f.formatNode(ctx.SyntaxBracketLc()))
		sb.WriteString(f.formatNode(ctx.ReplicationList()))
		sb.WriteString(f.formatNode(ctx.SyntaxBracketRc()))
	}

	// AND durable_writes = ...
	if ctx.DurableWrites() != nil {
		sb.WriteString("\n")
		sb.WriteString(indent)
		sb.WriteString("AND ")
		sb.WriteString(f.formatNode(ctx.DurableWrites()))
	}

	// AND tablets = ...
	if ctx.TabletsSpec() != nil {
		sb.WriteString("\n")
		sb.WriteString(indent)
		sb.WriteString("AND ")
		sb.WriteString(f.formatNode(ctx.TabletsSpec()))
	}

	sb.WriteString(";")
	return sb.String()
}

// formatAlterKeyspace formats an ALTER KEYSPACE statement
func (f *formatter) formatAlterKeyspace(ctx parser.IAlterKeyspaceContext) string {
	if ctx == nil {
		return f.formatTokenBased()
	}

	var sb strings.Builder
	indent := f.opts.IndentString

	// ALTER KEYSPACE
	sb.WriteString("ALTER KEYSPACE ")
	sb.WriteString(f.formatNode(ctx.Keyspace()))

	// WITH replication = {...}
	sb.WriteString("\n")
	sb.WriteString("WITH replication = ")
	if ctx.ReplicationList() != nil {
		sb.WriteString(f.formatNode(ctx.SyntaxBracketLc()))
		sb.WriteString(f.formatNode(ctx.ReplicationList()))
		sb.WriteString(f.formatNode(ctx.SyntaxBracketRc()))
	}

	// AND durable_writes = ...
	if ctx.DurableWrites() != nil {
		sb.WriteString("\n")
		sb.WriteString(indent)
		sb.WriteString("AND ")
		sb.WriteString(f.formatNode(ctx.DurableWrites()))
	}

	sb.WriteString(";")
	return sb.String()
}

// formatCreateMaterializedView formats a CREATE MATERIALIZED VIEW statement
func (f *formatter) formatCreateMaterializedView(ctx parser.ICreateMaterializedViewContext) string {
	if ctx == nil {
		return f.formatTokenBased()
	}

	var sb strings.Builder
	indent := f.opts.IndentString

	// CREATE MATERIALIZED VIEW [IF NOT EXISTS]
	sb.WriteString("CREATE MATERIALIZED VIEW")
	if ctx.IfNotExist() != nil {
		sb.WriteString(" IF NOT EXISTS")
	}

	// [keyspace.]view_name
	sb.WriteString(" ")
	if ks := ctx.Keyspace(); ks != nil {
		sb.WriteString(f.formatNode(ks))
		sb.WriteString(".")
	}
	sb.WriteString(f.formatNode(ctx.MaterializedView()))

	// AS
	sb.WriteString(" AS\n")

	// SELECT columns
	sb.WriteString(indent)
	sb.WriteString("SELECT ")
	if selElems := ctx.SelectElements(); selElems != nil {
		sb.WriteString(f.formatNode(selElems))
	}
	sb.WriteString("\n")

	// FROM table (FromSpec includes the FROM keyword)
	sb.WriteString(indent)
	if fromSpec := ctx.FromSpec(); fromSpec != nil {
		sb.WriteString(f.formatNode(fromSpec))
	}
	sb.WriteString("\n")

	// WHERE clause (MvWhereSpec includes the WHERE keyword)
	if mvWhere := ctx.MvWhereSpec(); mvWhere != nil {
		sb.WriteString(indent)
		sb.WriteString(f.formatNode(mvWhere))
		sb.WriteString("\n")
	}

	// PRIMARY KEY
	if pkElem := ctx.PrimaryKeyElement(); pkElem != nil {
		sb.WriteString(indent)
		sb.WriteString(f.formatPrimaryKeyElement(pkElem))
	}

	// WITH options
	if ctx.KwWith() != nil && ctx.MaterializedViewOptions() != nil {
		sb.WriteString("\n")
		sb.WriteString("WITH ")
		sb.WriteString(f.formatMaterializedViewOptions(ctx.MaterializedViewOptions()))
	}

	sb.WriteString(";")
	return sb.String()
}

// formatMaterializedViewOptions formats MV options
func (f *formatter) formatMaterializedViewOptions(ctx parser.IMaterializedViewOptionsContext) string {
	if ctx == nil {
		return ""
	}

	var parts []string

	// Handle CLUSTERING ORDER BY
	if clusteringOrder := ctx.ClusteringOrder(); clusteringOrder != nil {
		parts = append(parts, f.formatNode(clusteringOrder))
	}

	// Handle table options (which contains the other options)
	if tableOpts := ctx.TableOptions(); tableOpts != nil {
		parts = append(parts, f.collectTableOptions(tableOpts)...)
	}

	result := strings.Builder{}
	for i, part := range parts {
		if i > 0 {
			result.WriteString("\n")
			result.WriteString(f.opts.IndentString)
			result.WriteString("AND ")
		}
		result.WriteString(part)
	}

	return result.String()
}

// formatCreateIndex formats a CREATE INDEX statement
func (f *formatter) formatCreateIndex(ctx parser.ICreateIndexContext) string {
	if ctx == nil {
		return f.formatTokenBased()
	}

	var sb strings.Builder

	// CREATE [CUSTOM] INDEX [IF NOT EXISTS] [index_name]
	sb.WriteString("CREATE")
	if ctx.KwCustom() != nil {
		sb.WriteString(" CUSTOM")
	}
	sb.WriteString(" INDEX")
	if ctx.IfNotExist() != nil {
		sb.WriteString(" IF NOT EXISTS")
	}
	if indexName := ctx.OBJECT_NAME(); indexName != nil {
		sb.WriteString(" ")
		sb.WriteString(indexName.GetText())
	}

	// ON [keyspace.]table
	sb.WriteString("\n")
	sb.WriteString(f.opts.IndentString)
	sb.WriteString("ON ")
	if ks := ctx.Keyspace(); ks != nil {
		sb.WriteString(f.formatNode(ks))
		sb.WriteString(".")
	}
	sb.WriteString(f.formatNode(ctx.Table()))

	// (column) or (KEYS/VALUES/ENTRIES/FULL(column))
	sb.WriteString(" ")
	if bracket := ctx.SyntaxBracketLr(); bracket != nil {
		sb.WriteString(f.formatNode(bracket))
	}
	if idxColSpec := ctx.IndexColumnSpec(); idxColSpec != nil {
		sb.WriteString(f.formatNode(idxColSpec))
	}
	if bracket := ctx.SyntaxBracketRr(); bracket != nil {
		sb.WriteString(f.formatNode(bracket))
	}

	// USING 'class' [WITH OPTIONS = {...}]
	if indexUsing := ctx.IndexUsing(); indexUsing != nil {
		sb.WriteString("\n")
		sb.WriteString(f.opts.IndentString)
		sb.WriteString(f.formatNode(indexUsing))
	}

	sb.WriteString(";")
	return sb.String()
}

// formatSelect formats a SELECT statement
func (f *formatter) formatSelect(ctx parser.ISelect_Context) string {
	if ctx == nil {
		return f.formatTokenBased()
	}

	var sb strings.Builder
	indent := f.opts.IndentString

	// SELECT [DISTINCT] [JSON]
	sb.WriteString("SELECT")
	if ctx.DistinctSpec() != nil {
		sb.WriteString(" DISTINCT")
	}
	if ctx.KwJson() != nil {
		sb.WriteString(" JSON")
	}
	sb.WriteString(" ")

	// Select elements (columns)
	if selElems := ctx.SelectElements(); selElems != nil {
		sb.WriteString(f.formatNode(selElems))
	}
	sb.WriteString("\n")

	// FROM (FromSpec includes the FROM keyword)
	if fromSpec := ctx.FromSpec(); fromSpec != nil {
		sb.WriteString(f.formatNode(fromSpec))
	}

	// WHERE
	if whereSpec := ctx.WhereSpec(); whereSpec != nil {
		sb.WriteString("\n")
		sb.WriteString(f.formatNode(whereSpec))
	}

	// GROUP BY
	if groupBy := ctx.GroupBySpec(); groupBy != nil {
		sb.WriteString("\n")
		sb.WriteString(f.formatNode(groupBy))
	}

	// ORDER BY
	if orderSpec := ctx.OrderSpec(); orderSpec != nil {
		sb.WriteString("\n")
		sb.WriteString(f.formatNode(orderSpec))
	}

	// PER PARTITION LIMIT
	if perPartLimit := ctx.PerPartitionLimitSpec(); perPartLimit != nil {
		sb.WriteString("\n")
		sb.WriteString(f.formatNode(perPartLimit))
	}

	// LIMIT
	if limitSpec := ctx.LimitSpec(); limitSpec != nil {
		sb.WriteString("\n")
		sb.WriteString(f.formatNode(limitSpec))
	}

	// ALLOW FILTERING
	if allowFilter := ctx.AllowFilteringSpec(); allowFilter != nil {
		sb.WriteString("\n")
		sb.WriteString(f.formatNode(allowFilter))
	}

	// BYPASS CACHE
	if bypassCache := ctx.BypassCacheSpec(); bypassCache != nil {
		sb.WriteString("\n")
		sb.WriteString(f.formatNode(bypassCache))
	}

	// USING TIMEOUT
	if usingTimeout := ctx.UsingTimeoutSpec(); usingTimeout != nil {
		sb.WriteString("\n")
		sb.WriteString(f.formatNode(usingTimeout))
	}

	_ = indent // may use later for complex formatting
	sb.WriteString(";")
	return sb.String()
}

// formatInsert formats an INSERT statement
func (f *formatter) formatInsert(ctx parser.IInsertContext) string {
	if ctx == nil {
		return f.formatTokenBased()
	}

	var sb strings.Builder

	// INSERT INTO [keyspace.]table
	sb.WriteString("INSERT INTO ")
	if ks := ctx.Keyspace(); ks != nil {
		sb.WriteString(f.formatNode(ks))
		sb.WriteString(".")
	}
	sb.WriteString(f.formatNode(ctx.Table()))

	// (columns)
	if cols := ctx.InsertColumnSpec(); cols != nil {
		sb.WriteString(" ")
		sb.WriteString(f.formatNode(cols))
	}
	sb.WriteString("\n")

	// VALUES (values) - InsertValuesSpec includes the VALUES keyword
	if vals := ctx.InsertValuesSpec(); vals != nil {
		sb.WriteString(f.formatNode(vals))
	}

	// IF NOT EXISTS
	if ctx.IfNotExist() != nil {
		sb.WriteString("\n")
		sb.WriteString("IF NOT EXISTS")
	}

	// USING TTL/TIMESTAMP
	if usingClause := ctx.UsingTtlTimestamp(); usingClause != nil {
		sb.WriteString("\n")
		sb.WriteString(f.formatNode(usingClause))
	}

	sb.WriteString(";")
	return sb.String()
}

// formatUpdate formats an UPDATE statement
func (f *formatter) formatUpdate(ctx parser.IUpdateContext) string {
	if ctx == nil {
		return f.formatTokenBased()
	}

	var sb strings.Builder

	// UPDATE [keyspace.]table
	sb.WriteString("UPDATE ")
	if ks := ctx.Keyspace(); ks != nil {
		sb.WriteString(f.formatNode(ks))
		sb.WriteString(".")
	}
	sb.WriteString(f.formatNode(ctx.Table()))

	// USING TTL/TIMESTAMP (before SET)
	if usingClause := ctx.UsingTtlTimestamp(); usingClause != nil {
		sb.WriteString("\n")
		sb.WriteString(f.formatNode(usingClause))
	}

	// SET column = value, ...
	sb.WriteString("\n")
	sb.WriteString("SET ")
	if assignments := ctx.Assignments(); assignments != nil {
		sb.WriteString(f.formatNode(assignments))
	}

	// WHERE
	if whereSpec := ctx.WhereSpec(); whereSpec != nil {
		sb.WriteString("\n")
		sb.WriteString(f.formatNode(whereSpec))
	}

	// IF EXISTS / IF condition
	if ifSpec := ctx.IfSpec(); ifSpec != nil {
		sb.WriteString("\n")
		sb.WriteString(f.formatNode(ifSpec))
	}

	sb.WriteString(";")
	return sb.String()
}

// formatDelete formats a DELETE statement
func (f *formatter) formatDelete(ctx parser.IDelete_Context) string {
	if ctx == nil {
		return f.formatTokenBased()
	}

	var sb strings.Builder

	// DELETE [columns]
	sb.WriteString("DELETE")
	if delColSpec := ctx.DeleteColumnList(); delColSpec != nil {
		sb.WriteString(" ")
		sb.WriteString(f.formatNode(delColSpec))
	}

	// FROM [keyspace.]table (via FromSpec)
	sb.WriteString("\n")
	if fromSpec := ctx.FromSpec(); fromSpec != nil {
		sb.WriteString(f.formatNode(fromSpec))
	}

	// USING TIMESTAMP
	if usingClause := ctx.UsingTimestampSpec(); usingClause != nil {
		sb.WriteString("\n")
		sb.WriteString(f.formatNode(usingClause))
	}

	// WHERE
	if whereSpec := ctx.WhereSpec(); whereSpec != nil {
		sb.WriteString("\n")
		sb.WriteString(f.formatNode(whereSpec))
	}

	// IF EXISTS
	if ifExist := ctx.IfExist(); ifExist != nil {
		sb.WriteString("\n")
		sb.WriteString(f.formatNode(ifExist))
	}

	// IF condition
	if ifSpec := ctx.IfSpec(); ifSpec != nil {
		sb.WriteString("\n")
		sb.WriteString(f.formatNode(ifSpec))
	}

	sb.WriteString(";")
	return sb.String()
}

// formatBatch formats a BATCH statement
func (f *formatter) formatBatch(ctx parser.IBatchContext) string {
	if ctx == nil {
		return f.formatTokenBased()
	}

	var sb strings.Builder
	indent := f.opts.IndentString

	// BEGIN [UNLOGGED|COUNTER] BATCH
	sb.WriteString("BEGIN")
	if batchType := ctx.BatchType(); batchType != nil {
		sb.WriteString(" ")
		sb.WriteString(f.formatNode(batchType))
	}
	sb.WriteString(" BATCH")

	// USING TIMESTAMP
	if usingClause := ctx.UsingTimestampSpec(); usingClause != nil {
		sb.WriteString(" ")
		sb.WriteString(f.formatNode(usingClause))
	}
	sb.WriteString("\n")

	// Batch statements (inserts, updates, deletes)
	if batchList := ctx.BatchStatementList(); batchList != nil {
		for _, batchStmt := range batchList.AllBatchStatement() {
			sb.WriteString(indent)
			sb.WriteString(f.formatNode(batchStmt))
			sb.WriteString(";\n")
		}
	}

	sb.WriteString("APPLY BATCH;")
	return sb.String()
}

// formatNode formats any AST node by extracting its original text
func (f *formatter) formatNode(ctx antlr.ParserRuleContext) string {
	if ctx == nil {
		return ""
	}
	return f.getOriginalText(ctx)
}

// getOriginalTextRaw extracts the original text without any keyword uppercasing
func (f *formatter) getOriginalTextRaw(ctx antlr.ParserRuleContext) string {
	if ctx == nil {
		return ""
	}

	start := ctx.GetStart()
	stop := ctx.GetStop()

	if start == nil || stop == nil {
		return ""
	}

	// Get tokens in range and reconstruct with proper spacing (no uppercasing)
	f.tokens.Fill()
	allTokens := f.tokens.GetAllTokens()

	var parts []string
	var lastToken antlr.Token

	for _, token := range allTokens {
		if token.GetTokenIndex() >= start.GetTokenIndex() &&
			token.GetTokenIndex() <= stop.GetTokenIndex() &&
			token.GetChannel() == antlr.TokenDefaultChannel {

			text := token.GetText()

			if lastToken != nil && needsSpaceBetween(lastToken.GetText(), text) {
				parts = append(parts, " ")
			}

			parts = append(parts, text)
			lastToken = token
		}
	}

	return strings.Join(parts, "")
}

// getOriginalText extracts the original text from a parse tree node
func (f *formatter) getOriginalText(ctx antlr.ParserRuleContext) string {
	if ctx == nil {
		return ""
	}

	start := ctx.GetStart()
	stop := ctx.GetStop()

	if start == nil || stop == nil {
		return ""
	}

	// Get tokens in range and reconstruct with proper spacing
	f.tokens.Fill()
	allTokens := f.tokens.GetAllTokens()

	var parts []string
	var lastToken antlr.Token

	for _, token := range allTokens {
		if token.GetTokenIndex() >= start.GetTokenIndex() &&
			token.GetTokenIndex() <= stop.GetTokenIndex() &&
			token.GetChannel() == antlr.TokenDefaultChannel {

			text := token.GetText()
			if f.opts.UppercaseKeywords && isKeywordToken(token.GetTokenType()) {
				text = strings.ToUpper(text)
			}

			if lastToken != nil && needsSpaceBetween(lastToken.GetText(), text) {
				parts = append(parts, " ")
			}

			parts = append(parts, text)
			lastToken = token
		}
	}

	return strings.Join(parts, "")
}

// formatTokenBased is the fallback token-based formatter
func (f *formatter) formatTokenBased() string {
	f.output.Reset()
	f.indent = 0

	f.tokens.Fill()
	allTokens := f.tokens.GetAllTokens()

	var lastTokenText string
	var lastWasNewline bool

	for i, token := range allTokens {
		if token.GetChannel() == antlr.TokenHiddenChannel {
			continue
		}
		if token.GetTokenType() == antlr.TokenEOF {
			continue
		}

		text := token.GetText()
		upperText := strings.ToUpper(text)

		if f.opts.UppercaseKeywords && isKeywordToken(token.GetTokenType()) {
			text = upperText
		}

		if f.opts.Style == Pretty && !lastWasNewline {
			if shouldNewlineBefore, ok := newlineKeywords[upperText]; ok && shouldNewlineBefore {
				if f.output.Len() > 0 && lastTokenText != "(" {
					f.output.WriteString("\n")
					lastWasNewline = true
				}
			}
		}

		if f.output.Len() > 0 && !lastWasNewline {
			if needsSpaceBetween(lastTokenText, text) {
				f.output.WriteString(" ")
			}
		}

		f.output.WriteString(text)
		lastTokenText = text
		lastWasNewline = false

		if f.opts.Style == Pretty && text == ";" && i+1 < len(allTokens) {
			nextToken := allTokens[i+1]
			if nextToken.GetChannel() != antlr.TokenHiddenChannel {
				nextText := strings.ToUpper(nextToken.GetText())
				if nextText == "INSERT" || nextText == "UPDATE" || nextText == "DELETE" || nextText == "APPLY" {
					f.output.WriteString("\n")
					lastWasNewline = true
				}
			}
		}
	}

	result := strings.TrimSpace(f.output.String())

	if !strings.HasSuffix(result, ";") {
		result += ";"
	}

	return result
}

// Keywords that should start a new line in pretty mode
var newlineKeywords = map[string]bool{
	"FROM":    true,
	"WHERE":   true,
	"AND":     false,
	"OR":      false,
	"ORDER":   true,
	"GROUP":   true,
	"LIMIT":   true,
	"ALLOW":   true,
	"BYPASS":  true,
	"USING":   true,
	"SET":     true,
	"VALUES":  true,
	"IF":      true,
	"WITH":    true,
	"PRIMARY": false,
	"PER":     true,
	"BEGIN":   false,
	"APPLY":   true,
	"INSERT":  false,
	"UPDATE":  false,
	"DELETE":  false,
}

func needsSpaceBetween(prev, curr string) bool {
	if prev == "" || curr == "" {
		return false
	}

	if prev == "(" || prev == "[" || prev == "{" || prev == "<" {
		return false
	}

	if curr == ")" || curr == "]" || curr == "}" || curr == "," || curr == ";" || curr == "." || curr == ":" || curr == ">" {
		return false
	}

	if prev == "." {
		return false
	}

	if curr == "(" {
		upperPrev := strings.ToUpper(prev)
		keywordsWithSpace := map[string]bool{
			"IN": true, "KEY": true, "KEYS": true, "ENTRIES": true, "FULL": true,
			"VALUES": true, "PARTITION": true,
		}
		return keywordsWithSpace[upperPrev]
	}

	if curr == "<" {
		// No space before < in generic types (map<text, text>)
		return false
	}

	return true
}

func isKeywordToken(tokenType int) bool {
	return tokenType >= 1 && tokenType <= 180
}

// uppercaseDataType uppercases CQL type names while preserving UDT names
func uppercaseDataType(text string) string {
	// Common CQL types that should be uppercased
	typeNames := []string{
		"TEXT", "VARCHAR", "ASCII", "BLOB", "BOOLEAN", "BOOL",
		"INT", "BIGINT", "SMALLINT", "TINYINT", "VARINT",
		"FLOAT", "DOUBLE", "DECIMAL",
		"UUID", "TIMEUUID", "TIMESTAMP", "DATE", "TIME", "DURATION",
		"INET", "COUNTER",
		"LIST", "SET", "MAP", "TUPLE", "FROZEN",
	}

	result := text
	for _, typeName := range typeNames {
		// Case-insensitive replacement
		lower := strings.ToLower(typeName)
		// Replace only if it's a whole word (not part of identifier)
		result = replaceWholeWord(result, lower, typeName)
		result = replaceWholeWord(result, typeName, typeName)
	}

	return result
}

// replaceWholeWord replaces a word only when it appears as a whole word
func replaceWholeWord(s, old, new string) string {
	// Simple approach: just do case-insensitive replacement for type names
	result := s
	lowerS := strings.ToLower(s)
	lowerOld := strings.ToLower(old)

	idx := 0
	for {
		pos := strings.Index(strings.ToLower(result[idx:]), lowerOld)
		if pos == -1 {
			break
		}
		pos += idx

		// Check if it's a word boundary
		before := pos == 0 || !isIdentChar(result[pos-1])
		after := pos+len(old) >= len(result) || !isIdentChar(result[pos+len(old)])

		if before && after {
			result = result[:pos] + new + result[pos+len(old):]
		}
		idx = pos + len(new)
		if idx >= len(result) {
			break
		}
	}

	// Keep the original if no changes for case-sensitivity
	if strings.EqualFold(result, s) && result != s {
		// Preserve original casing if same word
	}

	_ = lowerS // silence unused warning
	return result
}

func isIdentChar(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9') || b == '_'
}
