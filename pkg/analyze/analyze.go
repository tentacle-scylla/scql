package analyze

import (
	"fmt"
	"strings"

	"github.com/tentacle-scylla/scql/pkg/schema"
	"github.com/tentacle-scylla/scql/pkg/types"
)

// Analyze performs full analysis of a CQL query with optional schema validation.
func Analyze(cql string, opts *AnalyzeOptions) *Result {
	if opts == nil {
		opts = DefaultOptions()
	}

	result := &Result{
		Query:        cql,
		SchemaErrors: make([]*SchemaError, 0),
		Warnings:     make([]*Warning, 0),
	}

	// Extract references from the query
	refs, stmtType, syntaxErrors := ExtractReferences(cql)
	result.References = refs
	result.Type = stmtType
	result.SyntaxErrors = syntaxErrors
	result.IsValid = len(syntaxErrors) == 0

	// If there are syntax errors, don't proceed with schema validation
	if !result.IsValid {
		return result
	}

	// Apply default keyspace if not specified
	if refs.Keyspace == "" && opts.DefaultKeyspace != "" {
		refs.Keyspace = opts.DefaultKeyspace
	}

	// Schema validation (only if schema is provided)
	if opts.Schema != nil && refs.Table != "" {
		validateSchema(result, opts)
	}

	// Function validation (always, doesn't require schema)
	if len(refs.FunctionCalls) > 0 {
		funcErrors := ValidateFunctionCalls(refs.FunctionCalls)
		result.SchemaErrors = append(result.SchemaErrors, funcErrors...)
	}

	// Generate warnings
	generateWarnings(result, opts)

	return result
}

// validateSchema validates the query references against the schema.
func validateSchema(result *Result, opts *AnalyzeOptions) {
	refs := result.References
	s := opts.Schema

	// Find the keyspace
	var ks *schema.Keyspace
	if refs.Keyspace != "" {
		ks = s.GetKeyspace(refs.Keyspace)
		if ks == nil {
			result.SchemaErrors = append(result.SchemaErrors, &SchemaError{
				Type:       ErrUnknownKeyspace,
				Message:    fmt.Sprintf("Keyspace '%s' does not exist", refs.Keyspace),
				Suggestion: suggestKeyspace(s, refs.Keyspace),
				Object:     refs.Keyspace,
			})
			return
		}
	} else if opts.DefaultKeyspace != "" {
		ks = s.GetKeyspace(opts.DefaultKeyspace)
	}

	if ks == nil {
		// No keyspace context - can't validate further
		return
	}

	// Find the table
	tbl := ks.GetTable(refs.Table)
	if tbl == nil {
		result.SchemaErrors = append(result.SchemaErrors, &SchemaError{
			Type:       ErrUnknownTable,
			Message:    fmt.Sprintf("Table '%s' does not exist in keyspace '%s'", refs.Table, ks.Name),
			Suggestion: suggestTable(ks, refs.Table),
			Object:     refs.Table,
		})
		return
	}

	// Validate columns
	for _, colName := range refs.Columns {
		if colName == "*" {
			continue
		}
		col := tbl.GetColumn(colName)
		if col == nil {
			result.SchemaErrors = append(result.SchemaErrors, &SchemaError{
				Type:       ErrUnknownColumn,
				Message:    fmt.Sprintf("Column '%s' not found in table '%s'", colName, tbl.Name),
				Suggestion: suggestColumn(tbl, colName),
				Object:     colName,
			})
		}
	}

	// Validate partition key usage for DML queries
	if result.Type.IsDML() && result.Type != types.StatementInsert {
		validatePartitionKey(result, tbl)
	}
}

// validatePartitionKey checks if the WHERE clause contains all partition key columns.
func validatePartitionKey(result *Result, tbl *schema.Table) {
	refs := result.References

	// For SELECT/UPDATE/DELETE, partition key should be in WHERE clause
	whereColSet := make(map[string]bool)
	for _, col := range refs.WhereColumns {
		whereColSet[strings.ToLower(col)] = true
	}

	// Check partition key columns
	missingPK := make([]string, 0)
	for _, pkCol := range tbl.PartitionKey {
		if !whereColSet[strings.ToLower(pkCol)] {
			missingPK = append(missingPK, pkCol)
		}
	}

	if len(missingPK) > 0 {
		// Missing partition key columns
		if len(missingPK) == len(tbl.PartitionKey) {
			result.Warnings = append(result.Warnings, &Warning{
				Type:       WarnMissingPartitionKey,
				Severity:   SeverityError,
				Message:    fmt.Sprintf("Query is missing partition key column(s): %s", strings.Join(missingPK, ", ")),
				Suggestion: "Add partition key columns to WHERE clause or use ALLOW FILTERING",
			})
		} else {
			result.Warnings = append(result.Warnings, &Warning{
				Type:       WarnMissingPartitionKey,
				Severity:   SeverityError,
				Message:    fmt.Sprintf("Query is missing partition key column(s): %s", strings.Join(missingPK, ", ")),
				Suggestion: "All partition key columns must be specified",
			})
		}
	}

	// Check clustering key order
	if len(tbl.ClusteringKey) > 0 && len(missingPK) == 0 {
		// Partition key is complete, now check clustering key prefix
		validateClusteringKeyOrder(result, tbl, whereColSet)
	}
}

// validateClusteringKeyOrder checks if clustering keys are used in proper prefix order.
func validateClusteringKeyOrder(result *Result, tbl *schema.Table, whereColSet map[string]bool) {
	// Clustering columns must be restricted in order
	// You can't skip a clustering column and restrict a later one
	foundGap := false
	for _, ckCol := range tbl.ClusteringKey {
		hasCol := whereColSet[strings.ToLower(ckCol)]
		if !hasCol {
			foundGap = true
		} else if foundGap {
			// Found a clustering column after a gap
			result.Warnings = append(result.Warnings, &Warning{
				Type:       WarnMissingClusteringKey,
				Severity:   SeverityWarning,
				Message:    fmt.Sprintf("Clustering column '%s' is used without preceding clustering columns", ckCol),
				Suggestion: "Include preceding clustering columns in WHERE clause or use ALLOW FILTERING",
			})
		}
	}
}

// generateWarnings generates warnings based on query characteristics.
func generateWarnings(result *Result, opts *AnalyzeOptions) {
	refs := result.References

	// Only for SELECT queries
	if result.Type != types.StatementSelect {
		return
	}

	// Warn about SELECT *
	if opts.WarnOnSelectStar {
		for _, col := range refs.SelectColumns {
			if col == "*" {
				result.Warnings = append(result.Warnings, &Warning{
					Type:       WarnSelectStar,
					Severity:   SeverityInfo,
					Message:    "SELECT * retrieves all columns",
					Suggestion: "Consider selecting only needed columns",
				})
				break
			}
		}
	}

	// Warn about missing LIMIT
	if opts.WarnOnNoLimit && refs.Limit < 0 {
		result.Warnings = append(result.Warnings, &Warning{
			Type:       WarnNoLimit,
			Severity:   SeverityInfo,
			Message:    "Query has no LIMIT clause",
			Suggestion: "Consider adding LIMIT to avoid fetching too many rows",
		})
	}

	// Warn about large LIMIT
	if opts.LargeLimitThreshold > 0 && refs.Limit > opts.LargeLimitThreshold {
		result.Warnings = append(result.Warnings, &Warning{
			Type:       WarnLargeLimit,
			Severity:   SeverityWarning,
			Message:    fmt.Sprintf("LIMIT %d may return too many rows", refs.Limit),
			Suggestion: fmt.Sprintf("Consider using a smaller limit (threshold: %d)", opts.LargeLimitThreshold),
		})
	}

	// Warn about missing WHERE clause
	if len(refs.WhereColumns) == 0 {
		result.Warnings = append(result.Warnings, &Warning{
			Type:       WarnNoWhereClause,
			Severity:   SeverityInfo,
			Message:    "Query has no WHERE clause - will scan entire table",
			Suggestion: "Add WHERE clause to filter results",
		})
	}

	// Warn about ALLOW FILTERING presence
	if refs.HasAllowFiltering {
		result.Warnings = append(result.Warnings, &Warning{
			Type:       WarnAllowFilteringPresent,
			Severity:   SeverityWarning,
			Message:    "Query uses ALLOW FILTERING which may be slow",
			Suggestion: "Consider restructuring query to avoid ALLOW FILTERING",
		})
	}
}

// Suggestion helpers using Levenshtein distance

func suggestKeyspace(s *schema.Schema, name string) string {
	if s == nil {
		return ""
	}
	names := s.KeyspaceNames()
	if suggestion := findClosest(name, names); suggestion != "" {
		return fmt.Sprintf("Did you mean '%s'?", suggestion)
	}
	return ""
}

func suggestTable(ks *schema.Keyspace, name string) string {
	if ks == nil {
		return ""
	}
	names := ks.TableNames()
	if suggestion := findClosest(name, names); suggestion != "" {
		return fmt.Sprintf("Did you mean '%s'?", suggestion)
	}
	return ""
}

func suggestColumn(tbl *schema.Table, name string) string {
	if tbl == nil {
		return ""
	}
	var names []string
	for _, col := range tbl.AllColumns() {
		names = append(names, col.Name)
	}
	if suggestion := findClosest(name, names); suggestion != "" {
		return fmt.Sprintf("Did you mean '%s'?", suggestion)
	}
	return ""
}

// findClosest finds the closest match using simple Levenshtein distance.
func findClosest(input string, candidates []string) string {
	input = strings.ToLower(input)
	bestMatch := ""
	bestDistance := 3 // Max distance threshold

	for _, candidate := range candidates {
		dist := levenshtein(input, strings.ToLower(candidate))
		if dist < bestDistance {
			bestDistance = dist
			bestMatch = candidate
		}
	}

	return bestMatch
}

// levenshtein calculates the Levenshtein distance between two strings.
func levenshtein(s1, s2 string) int {
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	// Create matrix
	d := make([][]int, len(s1)+1)
	for i := range d {
		d[i] = make([]int, len(s2)+1)
		d[i][0] = i
	}
	for j := range d[0] {
		d[0][j] = j
	}

	// Fill matrix
	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			cost := 1
			if s1[i-1] == s2[j-1] {
				cost = 0
			}
			d[i][j] = min(
				d[i-1][j]+1,      // deletion
				d[i][j-1]+1,      // insertion
				d[i-1][j-1]+cost, // substitution
			)
		}
	}

	return d[len(s1)][len(s2)]
}

func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}
