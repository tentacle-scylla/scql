package hover

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/pierre-borckmans/scql/pkg/schema"
)

// GetHoverInfo returns hover information for the token at the given position.
func GetHoverInfo(ctx *HoverContext) *HoverInfo {
	if ctx == nil || ctx.Query == "" {
		return nil
	}

	// Find the token at the cursor position
	token := FindTokenAtPosition(ctx.Query, ctx.Position)
	if token == nil || token.Text == "" {
		return nil
	}

	// Try to resolve hover info based on context
	return resolveHoverInfo(token, ctx)
}

// FindTokenAtPosition finds the token at the given cursor position.
func FindTokenAtPosition(query string, position int) *Token {
	if len(query) == 0 {
		return nil
	}
	if position < 0 {
		position = 0
	}

	// Handle cursor at end or past end - clamp to last character
	if position >= len(query) {
		position = len(query) - 1
	}

	// Check if cursor is on whitespace or delimiter - no token at this position
	ch := rune(query[position])
	if unicode.IsSpace(ch) {
		return nil
	}

	// Check if we're on a delimiter character (like parenthesis)
	if ch == '(' || ch == ')' || ch == ',' || ch == ';' {
		return &Token{
			Text:  string(ch),
			Start: position,
			End:   position + 1,
			Type:  TokenPunctuation,
		}
	}

	// Check if we're on an asterisk (wildcard)
	if ch == '*' {
		return &Token{
			Text:  "*",
			Start: position,
			End:   position + 1,
			Type:  TokenOperator,
		}
	}

	// Not on a token character - no token
	if !isTokenChar(ch) {
		return nil
	}

	// Find token boundaries
	start := position
	end := position

	// Expand backwards to find token start
	for start > 0 {
		r := rune(query[start-1])
		if !isTokenChar(r) {
			break
		}
		start--
	}

	// Expand forwards to find token end
	for end < len(query) {
		r := rune(query[end])
		if !isTokenChar(r) {
			break
		}
		end++
	}

	// No token found at position
	if start == end {
		return nil
	}

	text := query[start:end]
	tokenType := classifyToken(text, query, start)

	return &Token{
		Text:  text,
		Start: start,
		End:   end,
		Type:  tokenType,
	}
}

// isTokenChar returns true if the character can be part of a token.
func isTokenChar(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_'
}

// classifyToken determines the type of token based on context.
func classifyToken(text string, query string, start int) TokenType {
	upper := strings.ToUpper(text)
	lower := strings.ToLower(text)

	// Check if it's a function (followed by '(') - do this FIRST
	// This handles cases like token(), count(), writetime(), ttl() which are also keywords
	afterToken := strings.TrimSpace(query[start+len(text):])
	if len(afterToken) > 0 && afterToken[0] == '(' {
		if IsFunction(lower) {
			return TokenFunction
		}
		// Even if not a known function, if followed by ( it's likely a function call
		return TokenFunction
	}

	// Check if it's a keyword
	if IsKeyword(upper) {
		return TokenKeyword
	}

	// Check if it's a type (in CREATE TABLE context)
	if IsType(lower) {
		return TokenKeyword // Types are keywords in CQL
	}

	// Otherwise it's an identifier (table, column, keyspace name)
	return TokenIdentifier
}

// resolveHoverInfo generates hover content based on the token.
func resolveHoverInfo(token *Token, ctx *HoverContext) *HoverInfo {
	switch token.Type {
	case TokenKeyword:
		return resolveKeywordHover(token)
	case TokenFunction:
		return resolveFunctionHover(token)
	case TokenIdentifier:
		return resolveIdentifierHover(token, ctx)
	case TokenOperator:
		return resolveOperatorHover(token, ctx)
	default:
		return nil
	}
}

// resolveKeywordHover generates hover for a keyword.
func resolveKeywordHover(token *Token) *HoverInfo {
	// First check if it's a type (types take precedence as they're more specific)
	// This handles cases like TIMESTAMP which is both a keyword (USING TIMESTAMP)
	// and a type (column type)
	typeInfo := GetTypeInfo(token.Text)
	if typeInfo != nil {
		return &HoverInfo{
			Content: formatTypeHover(typeInfo),
			Range:   &Range{Start: token.Start, End: token.End},
			Kind:    HoverType,
			Name:    typeInfo.Name,
		}
	}

	info := GetKeywordInfo(token.Text)
	if info == nil {
		return nil
	}

	return &HoverInfo{
		Content: formatKeywordHover(info),
		Range:   &Range{Start: token.Start, End: token.End},
		Kind:    HoverKeyword,
		Name:    info.Name,
	}
}

// resolveFunctionHover generates hover for a function.
func resolveFunctionHover(token *Token) *HoverInfo {
	info := GetFunctionInfo(token.Text)
	if info == nil {
		// Unknown function - provide basic info
		return &HoverInfo{
			Content: fmt.Sprintf("**%s**\n\nFunction", token.Text),
			Range:   &Range{Start: token.Start, End: token.End},
			Kind:    HoverFunction,
			Name:    token.Text,
		}
	}

	return &HoverInfo{
		Content: formatFunctionHover(info),
		Range:   &Range{Start: token.Start, End: token.End},
		Kind:    HoverFunction,
		Name:    info.Name,
	}
}

// resolveIdentifierHover generates hover for an identifier (table, column, keyspace).
func resolveIdentifierHover(token *Token, ctx *HoverContext) *HoverInfo {
	if ctx.Schema == nil {
		return nil
	}

	// Extract query context to understand what type of identifier this might be
	queryCtx := extractQueryContext(ctx.Query, token.Start)

	// Try to resolve as column
	if queryCtx.tableName != "" {
		ks := queryCtx.keyspaceName
		if ks == "" {
			ks = ctx.DefaultKeyspace
		}
		if ks != "" {
			if col := findColumn(ctx.Schema, ks, queryCtx.tableName, token.Text); col != nil {
				return &HoverInfo{
					Content: formatColumnHover(col, queryCtx.tableName),
					Range:   &Range{Start: token.Start, End: token.End},
					Kind:    HoverColumn,
					Name:    col.Name,
				}
			}
		}
	}

	// Try to resolve as table
	if tbl := findTable(ctx.Schema, ctx.DefaultKeyspace, token.Text); tbl != nil {
		return &HoverInfo{
			Content: formatTableHover(tbl),
			Range:   &Range{Start: token.Start, End: token.End},
			Kind:    HoverTable,
			Name:    tbl.Name,
		}
	}

	// Try to resolve as keyspace
	if ks := ctx.Schema.GetKeyspace(token.Text); ks != nil {
		return &HoverInfo{
			Content: formatKeyspaceHover(ks),
			Range:   &Range{Start: token.Start, End: token.End},
			Kind:    HoverKeyspace,
			Name:    ks.Name,
		}
	}

	return nil
}

// resolveOperatorHover generates hover for an operator (like *).
func resolveOperatorHover(token *Token, _ *HoverContext) *HoverInfo {
	if token.Text == "*" {
		return &HoverInfo{
			Content: "**\\***\n\nSelects all columns from the table",
			Range:   &Range{Start: token.Start, End: token.End},
			Kind:    HoverOperator,
			Name:    "*",
		}
	}
	return nil
}

// queryContext holds extracted context from the query.
type queryContext struct {
	keyspaceName string
	tableName    string
}

// extractQueryContext extracts keyspace and table context from the query.
func extractQueryContext(query string, position int) *queryContext {
	ctx := &queryContext{}

	// Normalize query for analysis
	upper := strings.ToUpper(query)

	// Find FROM clause to get table name
	fromIdx := strings.LastIndex(upper[:min(position+50, len(upper))], "FROM ")
	if fromIdx == -1 {
		// Also check the rest of the query
		fromIdx = strings.Index(upper, "FROM ")
	}
	if fromIdx != -1 {
		afterFrom := strings.TrimSpace(query[fromIdx+5:])
		parts := strings.FieldsFunc(afterFrom, func(r rune) bool {
			return unicode.IsSpace(r) || r == ',' || r == ';' || r == '(' || r == ')'
		})
		if len(parts) > 0 {
			tableRef := parts[0]
			if dotIdx := strings.Index(tableRef, "."); dotIdx != -1 {
				ctx.keyspaceName = strings.ToLower(tableRef[:dotIdx])
				ctx.tableName = strings.ToLower(tableRef[dotIdx+1:])
			} else {
				ctx.tableName = strings.ToLower(tableRef)
			}
		}
	}

	// Check UPDATE clause
	if ctx.tableName == "" {
		updateIdx := strings.Index(upper, "UPDATE ")
		if updateIdx != -1 {
			afterUpdate := strings.TrimSpace(query[updateIdx+7:])
			parts := strings.FieldsFunc(afterUpdate, func(r rune) bool {
				return unicode.IsSpace(r) || r == ','
			})
			if len(parts) > 0 {
				tableRef := parts[0]
				if dotIdx := strings.Index(tableRef, "."); dotIdx != -1 {
					ctx.keyspaceName = strings.ToLower(tableRef[:dotIdx])
					ctx.tableName = strings.ToLower(tableRef[dotIdx+1:])
				} else {
					ctx.tableName = strings.ToLower(tableRef)
				}
			}
		}
	}

	// Check INSERT INTO clause
	if ctx.tableName == "" {
		intoIdx := strings.Index(upper, "INTO ")
		if intoIdx != -1 {
			afterInto := strings.TrimSpace(query[intoIdx+5:])
			parts := strings.FieldsFunc(afterInto, func(r rune) bool {
				return unicode.IsSpace(r) || r == '(' || r == ','
			})
			if len(parts) > 0 {
				tableRef := parts[0]
				if dotIdx := strings.Index(tableRef, "."); dotIdx != -1 {
					ctx.keyspaceName = strings.ToLower(tableRef[:dotIdx])
					ctx.tableName = strings.ToLower(tableRef[dotIdx+1:])
				} else {
					ctx.tableName = strings.ToLower(tableRef)
				}
			}
		}
	}

	return ctx
}

// findColumn finds a column in the schema.
func findColumn(s *schema.Schema, keyspace, table, column string) *schema.Column {
	ks := s.GetKeyspace(keyspace)
	if ks == nil {
		return nil
	}
	tbl := ks.GetTable(table)
	if tbl == nil {
		return nil
	}
	return tbl.GetColumn(column)
}

// findTable finds a table in the schema, searching default keyspace if needed.
func findTable(s *schema.Schema, defaultKs, table string) *schema.Table {
	// Try with dot notation first
	if dotIdx := strings.Index(table, "."); dotIdx != -1 {
		ks := s.GetKeyspace(table[:dotIdx])
		if ks != nil {
			return ks.GetTable(table[dotIdx+1:])
		}
	}

	// Try default keyspace
	if defaultKs != "" {
		ks := s.GetKeyspace(defaultKs)
		if ks != nil {
			if tbl := ks.GetTable(table); tbl != nil {
				return tbl
			}
		}
	}

	// Search all keyspaces
	for _, ksName := range s.KeyspaceNames() {
		ks := s.GetKeyspace(ksName)
		if ks != nil {
			if tbl := ks.GetTable(table); tbl != nil {
				return tbl
			}
		}
	}

	return nil
}

// formatKeywordHover formats hover content for a keyword.
func formatKeywordHover(info *KeywordInfo) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("**%s**\n\n", info.Name))
	sb.WriteString(info.Description)
	if info.Syntax != "" {
		sb.WriteString("\n\n```cql\n")
		sb.WriteString(info.Syntax)
		sb.WriteString("\n```")
	}
	return sb.String()
}

// formatFunctionHover formats hover content for a function.
func formatFunctionHover(info *FunctionInfo) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("**%s** → `%s`\n\n", info.Signature, info.ReturnType))
	sb.WriteString(info.Description)
	return sb.String()
}

// formatTypeHover formats hover content for a type.
func formatTypeHover(info *TypeInfo) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("**%s**\n\n", info.Name))
	sb.WriteString(info.Description)
	if info.Size != "" {
		sb.WriteString(fmt.Sprintf("\n\nSize: %s", info.Size))
	}
	return sb.String()
}

// formatColumnHover formats hover content for a column.
func formatColumnHover(col *schema.Column, tableName string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("**%s**: `%s`\n\n", col.Name, col.Type))

	var notes []string
	if col.IsPartitionKey {
		notes = append(notes, "partition key")
	}
	if col.IsClusteringKey {
		notes = append(notes, "clustering key")
	}
	if col.IsStatic {
		notes = append(notes, "static")
	}

	if len(notes) > 0 {
		sb.WriteString(fmt.Sprintf("*%s*\n\n", strings.Join(notes, ", ")))
	}

	sb.WriteString(fmt.Sprintf("Table: %s", tableName))
	return sb.String()
}

// formatTableHover formats hover content for a table.
func formatTableHover(tbl *schema.Table) string {
	var sb strings.Builder

	if tbl.Keyspace != "" {
		sb.WriteString(fmt.Sprintf("**%s.%s**\n\n", tbl.Keyspace, tbl.Name))
	} else {
		sb.WriteString(fmt.Sprintf("**%s**\n\n", tbl.Name))
	}

	// Column count
	colCount := len(tbl.AllColumns())
	sb.WriteString(fmt.Sprintf("%d column(s)\n\n", colCount))

	// Primary key info
	if len(tbl.PartitionKey) > 0 {
		sb.WriteString(fmt.Sprintf("**Partition key**: %s\n", strings.Join(tbl.PartitionKey, ", ")))
	}
	if len(tbl.ClusteringKey) > 0 {
		sb.WriteString(fmt.Sprintf("**Clustering key**: %s", strings.Join(tbl.ClusteringKey, ", ")))
	}

	return sb.String()
}

// formatKeyspaceHover formats hover content for a keyspace.
func formatKeyspaceHover(ks *schema.Keyspace) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("**%s** (keyspace)\n\n", ks.Name))

	tableCount := len(ks.TableNames())
	sb.WriteString(fmt.Sprintf("%d table(s)", tableCount))

	return sb.String()
}

// min returns the minimum of two integers.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
