package complete

import (
	"strings"
	"unicode"
)

// DetectContext analyzes the query and cursor position to determine completion context.
func DetectContext(query string, position int) *DetectedContext {
	// Ensure position is within bounds
	if position < 0 {
		position = 0
	}
	if position > len(query) {
		position = len(query)
	}

	// Get text before cursor
	textBefore := query[:position]

	// Find current token being typed
	tokenStart, prefix := findCurrentToken(textBefore)

	ctx := &DetectedContext{
		Type:       ContextUnknown,
		Prefix:     prefix,
		TokenStart: tokenStart,
		TokenEnd:   position,
	}

	// Normalize for analysis (uppercase, single spaces)
	normalized := normalizeForAnalysis(textBefore)

	// Detect context based on preceding text
	ctx.Type = detectContextType(normalized, prefix)

	// Extract keyspace/table context from FULL query (not just before cursor)
	// This allows column completion even when cursor is before FROM clause
	fullNormalized := normalizeForAnalysis(query)
	ctx.Keyspace, ctx.Table = extractTableContext(fullNormalized)

	// Extract column context for ContextAfterOperator
	if ctx.Type == ContextAfterOperator {
		ctx.Column = extractColumnBeforeOperator(normalized)
	}

	return ctx
}

// findCurrentToken finds the start of the current token and extracts the prefix.
func findCurrentToken(textBefore string) (int, string) {
	if len(textBefore) == 0 {
		return 0, ""
	}

	// Walk backwards to find token start
	i := len(textBefore) - 1

	// Skip any trailing whitespace (cursor is after whitespace)
	for i >= 0 && unicode.IsSpace(rune(textBefore[i])) {
		i--
	}

	if i < 0 {
		// All whitespace
		return len(textBefore), ""
	}

	// Check if we're right after whitespace (no prefix)
	if i < len(textBefore)-1 {
		// Cursor is after whitespace, no prefix
		return len(textBefore), ""
	}

	// Find start of current token
	tokenEnd := i + 1
	for i >= 0 {
		r := rune(textBefore[i])
		if unicode.IsSpace(r) || r == ',' || r == '(' || r == ')' || r == '.' || r == ';' {
			break
		}
		i--
	}

	tokenStart := i + 1
	prefix := textBefore[tokenStart:tokenEnd]

	return tokenStart, prefix
}

// normalizeForAnalysis prepares text for keyword detection.
func normalizeForAnalysis(text string) string {
	// Convert to uppercase
	text = strings.ToUpper(text)

	// Collapse multiple spaces
	var result strings.Builder
	prevSpace := false
	for _, r := range text {
		if unicode.IsSpace(r) {
			if !prevSpace {
				result.WriteRune(' ')
				prevSpace = true
			}
		} else {
			result.WriteRune(r)
			prevSpace = false
		}
	}

	return strings.TrimSpace(result.String())
}

// detectContextType determines the context type from normalized text.
func detectContextType(normalized string, prefix string) ContextType {
	// Remove the prefix from normalized text for matching
	normalized = strings.TrimSuffix(normalized, strings.ToUpper(prefix))
	normalized = strings.TrimSpace(normalized)

	// Empty or just whitespace = start of statement
	if normalized == "" {
		return ContextStatementStart
	}

	// Check for specific contexts (order matters - more specific first)

	// After dot (keyspace.table or table.column)
	if strings.HasSuffix(normalized, ".") {
		return ContextAfterDot
	}

	// After LIMIT - expects a number, no completions
	if normalized == "LIMIT" || strings.HasSuffix(normalized, " LIMIT") {
		return ContextAfterLimit
	}

	// After LIMIT <number> - only ALLOW FILTERING is valid
	if hasLimitWithValue(normalized) {
		return ContextAfterLimitValue
	}

	// After SELECT
	if normalized == "SELECT" || strings.HasSuffix(normalized, " SELECT") ||
		strings.HasSuffix(normalized, "(SELECT") {
		return ContextAfterSelect
	}

	// After FROM
	if normalized == "FROM" || strings.HasSuffix(normalized, " FROM") {
		return ContextAfterFrom
	}

	// After WHERE
	if normalized == "WHERE" || strings.HasSuffix(normalized, " WHERE") {
		return ContextAfterWhere
	}

	// After AND/OR in condition context
	if strings.HasSuffix(normalized, " AND") || strings.HasSuffix(normalized, " OR") {
		return ContextAfterAnd
	}

	// After operator (=, <, >, etc.)
	if strings.HasSuffix(normalized, " =") || strings.HasSuffix(normalized, " <") ||
		strings.HasSuffix(normalized, " >") || strings.HasSuffix(normalized, " <=") ||
		strings.HasSuffix(normalized, " >=") || strings.HasSuffix(normalized, " !=") ||
		strings.HasSuffix(normalized, " IN") {
		return ContextAfterOperator
	}

	// After INSERT INTO
	if strings.HasSuffix(normalized, "INSERT INTO") || strings.HasSuffix(normalized, " INTO") {
		// Check if we're in INSERT context
		if strings.Contains(normalized, "INSERT") {
			return ContextAfterInsertInto
		}
	}

	// After UPDATE
	if normalized == "UPDATE" || strings.HasSuffix(normalized, " UPDATE") {
		return ContextAfterUpdate
	}

	// After SET (in UPDATE)
	if strings.HasSuffix(normalized, " SET") {
		return ContextAfterSet
	}

	// After DELETE
	if normalized == "DELETE" || strings.HasSuffix(normalized, " DELETE") {
		return ContextAfterDelete
	}

	// After ORDER BY
	if strings.HasSuffix(normalized, "ORDER BY") {
		return ContextAfterOrderBy
	}

	// After GROUP BY
	if strings.HasSuffix(normalized, "GROUP BY") {
		return ContextAfterGroupBy
	}

	// After CREATE
	if normalized == "CREATE" || strings.HasSuffix(normalized, " CREATE") {
		return ContextAfterCreate
	}

	// After ALTER
	if normalized == "ALTER" || strings.HasSuffix(normalized, " ALTER") {
		return ContextAfterAlter
	}

	// After DROP
	if normalized == "DROP" || strings.HasSuffix(normalized, " DROP") {
		return ContextAfterDrop
	}

	// After USE
	if normalized == "USE" || strings.HasSuffix(normalized, " USE") {
		return ContextAfterUse
	}

	// After DESCRIBE/DESC
	if normalized == "DESCRIBE" || normalized == "DESC" ||
		strings.HasSuffix(normalized, " DESCRIBE") || strings.HasSuffix(normalized, " DESC") {
		return ContextAfterDescribe
	}

	// After PRUNE
	if normalized == "PRUNE" || strings.HasSuffix(normalized, " PRUNE") {
		return ContextAfterPrune
	}

	// After VALUES
	if strings.HasSuffix(normalized, " VALUES") || strings.HasSuffix(normalized, " VALUES(") {
		return ContextAfterValues
	}

	// Inside column list (after opening paren in INSERT)
	if strings.Contains(normalized, "INSERT INTO") {
		// Count parens to see if we're inside column list
		openCount := strings.Count(normalized, "(")
		closeCount := strings.Count(normalized, ")")
		if openCount > closeCount && !strings.Contains(normalized, "VALUES") {
			return ContextInColumnList
		}
	}

	// In SELECT column list (after SELECT, before FROM) - suggest columns, functions, and FROM
	// This handles: "SELECT |", "SELECT col|", "SELECT col,|", "SELECT col, col2|"
	if strings.HasPrefix(normalized, "SELECT") && !strings.Contains(normalized, " FROM") {
		return ContextAfterSelect
	}

	// After FROM with table already specified, suggest clauses
	if containsFromWithTable(normalized) && !strings.HasSuffix(normalized, ")") {
		// Check if we're in SELECT context
		if strings.HasPrefix(normalized, "SELECT") {
			// Check if we're in WHERE clause with incomplete condition (column without operator)
			if hasWhereWithIncompleteCondition(normalized) {
				return ContextAfterColumn
			}
			// After WHERE with a complete condition
			if hasCompleteWhereCondition(normalized) {
				return ContextAfterAnd
			}
			// Ready for WHERE, ORDER BY, LIMIT etc.
			return ContextAfterSelectTable
		}
	}

	// After UPDATE table SET col = value
	if strings.HasPrefix(normalized, "UPDATE") && strings.Contains(normalized, " SET ") {
		// Check if we have a complete SET assignment (col = value)
		if hasCompleteSetAssignment(normalized) {
			return ContextAfterUpdateSet
		}
	}

	// After DELETE FROM table (ready for WHERE)
	if strings.HasPrefix(normalized, "DELETE") && strings.Contains(normalized, " FROM ") {
		// Check if we have a table after FROM
		if containsDeleteFromWithTable(normalized) {
			return ContextAfterDeleteFrom
		}
	}

	return ContextUnknown
}

// containsFromWithTable checks if the query has FROM followed by a table name.
func containsFromWithTable(normalized string) bool {
	idx := strings.LastIndex(normalized, "FROM ")
	if idx == -1 {
		return false
	}
	afterFrom := strings.TrimSpace(normalized[idx+5:])
	// Check if there's at least one word after FROM
	parts := strings.Fields(afterFrom)
	return len(parts) > 0 && !isKeyword(parts[0])
}

// hasWhereWithIncompleteCondition checks if WHERE has a column but no operator yet.
func hasWhereWithIncompleteCondition(normalized string) bool {
	// Find WHERE or AND position
	whereIdx := strings.LastIndex(normalized, " WHERE ")
	andIdx := strings.LastIndex(normalized, " AND ")

	// Use the later one
	startIdx := whereIdx
	startLen := 7 // len(" WHERE ")
	if andIdx > whereIdx {
		startIdx = andIdx
		startLen = 5 // len(" AND ")
	}

	if startIdx == -1 {
		return false
	}

	afterKeyword := strings.TrimSpace(normalized[startIdx+startLen:])
	if afterKeyword == "" {
		return false
	}

	// Check if we have a word (column) but no operator
	parts := strings.Fields(afterKeyword)
	if len(parts) == 0 {
		return false
	}

	// If we have exactly one word and it's not a keyword, it's a column needing operator
	if len(parts) == 1 {
		word := parts[0]
		// Not a keyword and not ending with operator chars
		if !isKeyword(word) && !strings.HasSuffix(word, "=") &&
			!strings.HasSuffix(word, "<") && !strings.HasSuffix(word, ">") {
			return true
		}
	}

	// Check if the last word might be a column (no operator after it)
	lastWord := parts[len(parts)-1]
	if len(parts) >= 1 {
		// Check if there's no operator in the clause yet
		hasOperator := false
		operators := []string{" = ", " != ", " < ", " > ", " <= ", " >= ", " IN "}
		for _, op := range operators {
			if strings.Contains(afterKeyword, op) {
				hasOperator = true
				break
			}
		}
		// Also check for = without spaces
		if strings.Contains(afterKeyword, "=") {
			hasOperator = true
		}
		if !hasOperator && !isKeyword(lastWord) {
			return true
		}
	}

	return false
}

// hasCompleteWhereCondition checks if there's a complete WHERE condition (col op value).
func hasCompleteWhereCondition(normalized string) bool {
	whereIdx := strings.LastIndex(normalized, " WHERE ")
	if whereIdx == -1 {
		return false
	}
	afterWhere := strings.TrimSpace(normalized[whereIdx+7:])
	// A complete condition has at least: column operator value
	// Look for an operator followed by something
	operators := []string{" = ", " != ", " < ", " > ", " <= ", " >= ", " IN "}
	for _, op := range operators {
		if opIdx := strings.Index(afterWhere, op); opIdx != -1 {
			// Check if there's something after the operator
			afterOp := strings.TrimSpace(afterWhere[opIdx+len(op):])
			if len(afterOp) > 0 {
				return true
			}
		}
	}
	return false
}

// hasCompleteSetAssignment checks if UPDATE has a complete SET col = value.
func hasCompleteSetAssignment(normalized string) bool {
	setIdx := strings.Index(normalized, " SET ")
	if setIdx == -1 {
		return false
	}
	afterSet := strings.TrimSpace(normalized[setIdx+5:])
	// A complete SET has at least: column = value
	// Look for = followed by something
	eqIdx := strings.Index(afterSet, " = ")
	if eqIdx != -1 {
		afterEq := strings.TrimSpace(afterSet[eqIdx+3:])
		if len(afterEq) > 0 {
			return true
		}
	}
	// Also check for = without spaces (col=value)
	eqIdx = strings.Index(afterSet, "=")
	if eqIdx != -1 && eqIdx < len(afterSet)-1 {
		afterEq := strings.TrimSpace(afterSet[eqIdx+1:])
		if len(afterEq) > 0 {
			return true
		}
	}
	return false
}

// containsDeleteFromWithTable checks if DELETE FROM has a table name.
func containsDeleteFromWithTable(normalized string) bool {
	fromIdx := strings.Index(normalized, " FROM ")
	if fromIdx == -1 {
		return false
	}
	afterFrom := strings.TrimSpace(normalized[fromIdx+6:])
	parts := strings.Fields(afterFrom)
	return len(parts) > 0 && !isKeyword(parts[0])
}

// isKeyword checks if a word is a CQL keyword.
func isKeyword(word string) bool {
	keywords := []string{
		"SELECT", "FROM", "WHERE", "AND", "OR", "ORDER", "BY", "GROUP",
		"LIMIT", "INSERT", "INTO", "VALUES", "UPDATE", "SET", "DELETE",
		"CREATE", "ALTER", "DROP", "IF", "EXISTS", "NOT", "ALLOW", "FILTERING",
		"USING", "TTL", "TIMESTAMP", "ASC", "DESC", "IN", "CONTAINS", "KEY",
	}
	word = strings.ToUpper(word)
	for _, kw := range keywords {
		if word == kw {
			return true
		}
	}
	return false
}

// extractTableContext extracts keyspace and table from the query.
func extractTableContext(normalized string) (keyspace string, table string) {
	// Look for FROM clause
	fromIdx := strings.LastIndex(normalized, "FROM ")
	if fromIdx != -1 {
		afterFrom := normalized[fromIdx+5:]
		parts := strings.Fields(afterFrom)
		if len(parts) > 0 {
			tableRef := parts[0]
			// Remove trailing punctuation
			tableRef = strings.TrimRight(tableRef, ",;()")
			return parseTableRef(tableRef)
		}
	}

	// Look for UPDATE clause
	updateIdx := strings.Index(normalized, "UPDATE ")
	if updateIdx != -1 {
		afterUpdate := normalized[updateIdx+7:]
		parts := strings.Fields(afterUpdate)
		if len(parts) > 0 {
			tableRef := parts[0]
			tableRef = strings.TrimRight(tableRef, ",;()")
			return parseTableRef(tableRef)
		}
	}

	// Look for INSERT INTO
	intoIdx := strings.Index(normalized, "INTO ")
	if intoIdx != -1 {
		afterInto := normalized[intoIdx+5:]
		parts := strings.Fields(afterInto)
		if len(parts) > 0 {
			tableRef := parts[0]
			tableRef = strings.TrimRight(tableRef, ",;()")
			return parseTableRef(tableRef)
		}
	}

	return "", ""
}

// parseTableRef parses "keyspace.table" or just "table".
func parseTableRef(ref string) (keyspace string, table string) {
	ref = strings.Trim(ref, "\"")
	if idx := strings.Index(ref, "."); idx != -1 {
		return strings.ToLower(ref[:idx]), strings.ToLower(ref[idx+1:])
	}
	return "", strings.ToLower(ref)
}

// extractColumnBeforeOperator extracts the column name before an operator.
// For "WHERE build_hour =" returns "build_hour"
func extractColumnBeforeOperator(normalized string) string {
	// Find the last operator
	operators := []string{" = ", " != ", " < ", " > ", " <= ", " >= ", " IN "}
	lastOpIdx := -1

	for _, op := range operators {
		if idx := strings.LastIndex(normalized, op); idx > lastOpIdx {
			lastOpIdx = idx
		}
	}

	// Also check for operators at end without trailing space
	endOps := []string{" =", " !=", " <", " >", " <=", " >=", " IN"}
	for _, op := range endOps {
		if strings.HasSuffix(normalized, op) {
			idx := len(normalized) - len(op)
			if idx > lastOpIdx {
				lastOpIdx = idx
			}
		}
	}

	if lastOpIdx == -1 {
		return ""
	}

	// Get text before the operator
	beforeOp := strings.TrimSpace(normalized[:lastOpIdx])
	if beforeOp == "" {
		return ""
	}

	// Find the column name - it's the last word before the operator
	// But we need to handle "WHERE col" or "AND col" patterns
	parts := strings.Fields(beforeOp)
	if len(parts) == 0 {
		return ""
	}

	// The column is the last word
	column := parts[len(parts)-1]

	// Skip if it's a keyword
	if isKeyword(column) {
		return ""
	}

	return strings.ToLower(column)
}

// hasLimitWithValue checks if the query has LIMIT followed by a number.
func hasLimitWithValue(normalized string) bool {
	limitIdx := strings.LastIndex(normalized, " LIMIT ")
	if limitIdx == -1 {
		return false
	}
	afterLimit := strings.TrimSpace(normalized[limitIdx+7:])
	if afterLimit == "" {
		return false
	}
	// Check if what follows LIMIT looks like a number
	parts := strings.Fields(afterLimit)
	if len(parts) == 0 {
		return false
	}
	// Check if first token after LIMIT is a number or ?
	firstToken := parts[0]
	if firstToken == "?" {
		return true
	}
	// Check if it's a number
	for _, r := range firstToken {
		if r < '0' || r > '9' {
			return false
		}
	}
	return len(firstToken) > 0
}
