// Package analyze provides schema-aware analysis of CQL queries.
// It can validate queries against a schema, extract referenced objects,
// and detect potential issues like missing partition keys.
package analyze

import (
	"github.com/pierre-borckmans/scql/pkg/schema"
	"github.com/pierre-borckmans/scql/pkg/types"
)

// Result contains the result of analyzing a CQL query.
type Result struct {
	// Query is the original CQL query
	Query string

	// Type is the statement type (SELECT, INSERT, UPDATE, DELETE, etc.)
	Type types.StatementType

	// IsValid is true if the query is syntactically valid
	IsValid bool

	// SyntaxErrors contains any syntax errors from parsing
	SyntaxErrors types.Errors

	// References contains all schema objects referenced in the query
	References *References

	// SchemaErrors contains errors from schema validation (unknown tables, columns, etc.)
	SchemaErrors []*SchemaError

	// Warnings contains non-fatal issues (missing PK, ALLOW FILTERING needed, etc.)
	Warnings []*Warning
}

// References contains all schema objects referenced in a query.
type References struct {
	// Keyspace is the target keyspace (explicit or from USE)
	Keyspace string

	// Table is the target table name
	Table string

	// Columns are all column names referenced in the query
	Columns []string

	// SelectColumns are columns in the SELECT clause (for SELECT queries)
	SelectColumns []string

	// WhereColumns are columns in the WHERE clause
	WhereColumns []string

	// UpdateColumns are columns being SET (for UPDATE queries)
	UpdateColumns []string

	// InsertColumns are columns in the INSERT column list
	InsertColumns []string

	// OrderByColumns are columns in ORDER BY clause
	OrderByColumns []string

	// Functions are function names in the query (for backward compat)
	Functions []string

	// FunctionCalls contains detailed info about each function call
	FunctionCalls []*FunctionCall

	// HasAllowFiltering is true if ALLOW FILTERING is present
	HasAllowFiltering bool

	// Limit is the LIMIT value if present, -1 otherwise
	Limit int
}

// FunctionCall represents a function call in a query with argument details.
type FunctionCall struct {
	// Name is the function name (lowercase)
	Name string

	// ArgCount is the number of arguments passed
	ArgCount int

	// HasStar is true if the argument is * (e.g., count(*))
	HasStar bool

	// Position is the location in the query
	Position *Position
}

// SchemaError represents an error from schema validation.
type SchemaError struct {
	Type       SchemaErrorType
	Message    string
	Suggestion string
	Object     string // The object name that caused the error (keyspace, table, column)
	Position   *Position
}

// SchemaErrorType identifies the kind of schema error.
type SchemaErrorType string

const (
	ErrUnknownKeyspace       SchemaErrorType = "unknown_keyspace"
	ErrUnknownTable          SchemaErrorType = "unknown_table"
	ErrUnknownColumn         SchemaErrorType = "unknown_column"
	ErrUnknownFunction       SchemaErrorType = "unknown_function"
	ErrTypeMismatch          SchemaErrorType = "type_mismatch"
	ErrFunctionArgCount      SchemaErrorType = "function_arg_count"
	ErrFunctionArgCountRange SchemaErrorType = "function_arg_count_range"
)

// Warning represents a non-fatal issue with the query.
type Warning struct {
	Type       WarningType
	Severity   Severity
	Message    string
	Suggestion string
	Position   *Position
}

// WarningType identifies the kind of warning.
type WarningType string

const (
	WarnMissingPartitionKey   WarningType = "missing_partition_key"
	WarnMissingClusteringKey  WarningType = "missing_clustering_key"
	WarnAllowFilteringNeeded  WarningType = "allow_filtering_needed"
	WarnAllowFilteringPresent WarningType = "allow_filtering_present"
	WarnNoWhereClause         WarningType = "no_where_clause"
	WarnNoLimit               WarningType = "no_limit"
	WarnLargeLimit            WarningType = "large_limit"
	WarnSelectStar            WarningType = "select_star"
)

// Severity indicates how serious a warning is.
type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
	SeverityInfo    Severity = "info"
)

// Position represents a location in the query.
type Position struct {
	Line   int
	Column int
	Offset int
}

// HasSchemaErrors returns true if there are any schema validation errors.
func (r *Result) HasSchemaErrors() bool {
	return len(r.SchemaErrors) > 0
}

// HasWarnings returns true if there are any warnings.
func (r *Result) HasWarnings() bool {
	return len(r.Warnings) > 0
}

// HasErrors returns true if there are any errors (syntax or schema).
func (r *Result) HasErrors() bool {
	return len(r.SyntaxErrors) > 0 || len(r.SchemaErrors) > 0
}

// AllErrors returns all errors (syntax + schema) as a combined list.
func (r *Result) AllErrors() []string {
	var errs []string
	for _, e := range r.SyntaxErrors {
		errs = append(errs, e.DisplayMessage())
	}
	for _, e := range r.SchemaErrors {
		errs = append(errs, e.Message)
	}
	return errs
}

// WarningsOfType returns all warnings of a specific type.
func (r *Result) WarningsOfType(t WarningType) []*Warning {
	var result []*Warning
	for _, w := range r.Warnings {
		if w.Type == t {
			result = append(result, w)
		}
	}
	return result
}

// ErrorsOfType returns all schema errors of a specific type.
func (r *Result) ErrorsOfType(t SchemaErrorType) []*SchemaError {
	var result []*SchemaError
	for _, e := range r.SchemaErrors {
		if e.Type == t {
			result = append(result, e)
		}
	}
	return result
}

// NewReferences creates a new empty References.
func NewReferences() *References {
	return &References{
		Columns:        make([]string, 0),
		SelectColumns:  make([]string, 0),
		WhereColumns:   make([]string, 0),
		UpdateColumns:  make([]string, 0),
		InsertColumns:  make([]string, 0),
		OrderByColumns: make([]string, 0),
		Functions:      make([]string, 0),
		FunctionCalls:  make([]*FunctionCall, 0),
		Limit:          -1,
	}
}

// AnalyzeOptions configures analysis behavior.
type AnalyzeOptions struct {
	// Schema is the schema to validate against (optional)
	Schema *schema.Schema

	// DefaultKeyspace is used when no keyspace is specified in the query
	DefaultKeyspace string

	// WarnOnSelectStar warns about SELECT * queries
	WarnOnSelectStar bool

	// WarnOnNoLimit warns about SELECT queries without LIMIT
	WarnOnNoLimit bool

	// LargeLimitThreshold triggers a warning when LIMIT exceeds this value (0 = disabled)
	LargeLimitThreshold int
}

// DefaultOptions returns default analysis options.
func DefaultOptions() *AnalyzeOptions {
	return &AnalyzeOptions{
		WarnOnSelectStar:    false,
		WarnOnNoLimit:       false,
		LargeLimitThreshold: 0,
	}
}
