// Package scql provides CQL parsing, linting, and formatting capabilities
// for Cassandra and ScyllaDB.
//
// This is a convenience package that re-exports the main types and functions
// from the sub-packages. For more control, import the sub-packages directly:
//
//   - github.com/pierre-borckmans/scql/pkg/parse    - Parsing CQL statements
//   - github.com/pierre-borckmans/scql/pkg/format   - Formatting CQL statements
//   - github.com/pierre-borckmans/scql/pkg/lint     - Linting CQL statements
//   - github.com/pierre-borckmans/scql/pkg/types    - Common types (Error, StatementType)
//   - github.com/pierre-borckmans/scql/pkg/schema   - Schema types
//   - github.com/pierre-borckmans/scql/pkg/complete - Auto-completion
//   - github.com/pierre-borckmans/scql/pkg/hover    - Hover information
package scql

import (
	"github.com/pierre-borckmans/scql/pkg/analyze"
	"github.com/pierre-borckmans/scql/pkg/complete"
	"github.com/pierre-borckmans/scql/pkg/format"
	"github.com/pierre-borckmans/scql/pkg/hover"
	"github.com/pierre-borckmans/scql/pkg/lint"
	"github.com/pierre-borckmans/scql/pkg/parse"
	"github.com/pierre-borckmans/scql/pkg/schema"
	"github.com/pierre-borckmans/scql/pkg/tokenize"
	"github.com/pierre-borckmans/scql/pkg/types"
)

// Re-export types
type (
	// Error represents a parsing or validation error with position information
	Error = types.Error

	// Errors is a collection of Error pointers
	Errors = types.Errors

	// StatementType represents the type of CQL statement
	StatementType = types.StatementType

	// ParseResult contains the result of parsing a CQL statement
	ParseResult = parse.Result

	// LintResult contains detailed lint results for a statement
	LintResult = lint.Result

	// FormatStyle defines the formatting style for CQL output
	FormatStyle = format.Style

	// FormatOptions configures the formatter behavior
	FormatOptions = format.Options

	// Schema represents a CQL schema (keyspaces, tables, columns, etc.)
	Schema = schema.Schema

	// Keyspace represents a keyspace in the schema
	Keyspace = schema.Keyspace

	// Table represents a table in the schema
	Table = schema.Table

	// Column represents a column in a table
	Column = schema.Column

	// CompletionItem represents a single completion suggestion
	CompletionItem = complete.CompletionItem

	// CompletionContext contains all information needed to generate completions
	CompletionContext = complete.CompletionContext

	// CompletionKind identifies the type of completion item
	CompletionKind = complete.CompletionKind

	// HoverInfo contains information about a token at a position
	HoverInfo = hover.HoverInfo

	// HoverContext provides context for hover resolution
	HoverContext = hover.HoverContext

	// HoverKind identifies the type of hover target
	HoverKind = hover.HoverKind

	// AnalyzeResult contains the result of analyzing a CQL query
	AnalyzeResult = analyze.Result

	// AnalyzeOptions configures analysis behavior
	AnalyzeOptions = analyze.AnalyzeOptions

	// SchemaError represents an error from schema validation
	SchemaError = analyze.SchemaError

	// FunctionCall represents a function call in a query with argument details
	FunctionCall = analyze.FunctionCall

	// Token represents a single token for syntax highlighting
	Token = tokenize.Token

	// TokenType identifies the semantic type of a token
	TokenType = tokenize.TokenType

	// TokenContext provides semantic information for enhanced tokenization
	TokenContext = tokenize.Context
)

// Re-export statement type constants
const (
	StatementUnknown               = types.StatementUnknown
	StatementSelect                = types.StatementSelect
	StatementInsert                = types.StatementInsert
	StatementUpdate                = types.StatementUpdate
	StatementDelete                = types.StatementDelete
	StatementBatch                 = types.StatementBatch
	StatementCreateKeyspace        = types.StatementCreateKeyspace
	StatementAlterKeyspace         = types.StatementAlterKeyspace
	StatementDropKeyspace          = types.StatementDropKeyspace
	StatementCreateTable           = types.StatementCreateTable
	StatementAlterTable            = types.StatementAlterTable
	StatementDropTable             = types.StatementDropTable
	StatementTruncate              = types.StatementTruncate
	StatementCreateIndex           = types.StatementCreateIndex
	StatementDropIndex             = types.StatementDropIndex
	StatementCreateMaterializedView = types.StatementCreateMaterializedView
	StatementAlterMaterializedView  = types.StatementAlterMaterializedView
	StatementDropMaterializedView   = types.StatementDropMaterializedView
	StatementCreateType            = types.StatementCreateType
	StatementAlterType             = types.StatementAlterType
	StatementDropType              = types.StatementDropType
	StatementCreateFunction        = types.StatementCreateFunction
	StatementDropFunction          = types.StatementDropFunction
	StatementCreateAggregate       = types.StatementCreateAggregate
	StatementDropAggregate         = types.StatementDropAggregate
	StatementCreateTrigger         = types.StatementCreateTrigger
	StatementDropTrigger           = types.StatementDropTrigger
	StatementCreateRole            = types.StatementCreateRole
	StatementAlterRole             = types.StatementAlterRole
	StatementDropRole              = types.StatementDropRole
	StatementCreateUser            = types.StatementCreateUser
	StatementAlterUser             = types.StatementAlterUser
	StatementDropUser              = types.StatementDropUser
	StatementGrant                 = types.StatementGrant
	StatementRevoke                = types.StatementRevoke
	StatementListRoles             = types.StatementListRoles
	StatementListPermissions       = types.StatementListPermissions
	StatementUse                   = types.StatementUse
	StatementPruneMaterializedView = types.StatementPruneMaterializedView
)

// Re-export format style constants
const (
	FormatCompact = format.Compact
	FormatPretty  = format.Pretty
)

// Re-export token type constants
const (
	TokenKeyword       = tokenize.TokenKeyword
	TokenFunction      = tokenize.TokenFunction
	TokenTypeToken     = tokenize.TokenType_
	TokenString        = tokenize.TokenString
	TokenNumber        = tokenize.TokenNumber
	TokenComment       = tokenize.TokenComment
	TokenIdentifier    = tokenize.TokenIdentifier
	TokenOperator      = tokenize.TokenOperator
	TokenPunctuation   = tokenize.TokenPunctuation
	TokenPartitionKey  = tokenize.TokenPartitionKey
	TokenClusteringKey = tokenize.TokenClusteringKey
	TokenColumn        = tokenize.TokenColumn
	TokenPlaceholder   = tokenize.TokenPlaceholder
)

// Parse parses a single CQL statement
func Parse(input string) *ParseResult {
	return parse.Parse(input)
}

// ParseMultiple parses multiple CQL statements separated by semicolons
func ParseMultiple(input string) []*ParseResult {
	return parse.Multiple(input)
}

// IsValid returns true if the CQL input is syntactically valid
func IsValid(input string) bool {
	return parse.IsValid(input)
}

// Lint validates CQL and returns any errors found
func Lint(input string) Errors {
	return lint.Check(input)
}

// LintMultiple validates multiple CQL statements and returns all errors
func LintMultiple(input string) Errors {
	return lint.CheckMultiple(input)
}

// Analyze performs detailed analysis on a CQL statement
func Analyze(input string) *LintResult {
	return lint.Analyze(input)
}

// AnalyzeMultiple performs detailed analysis on multiple CQL statements
func AnalyzeMultiple(input string) []*LintResult {
	return lint.AnalyzeMultiple(input)
}

// Format formats a parsed CQL statement according to the given options
func Format(result *ParseResult, opts FormatOptions) string {
	return format.Format(result, opts)
}

// FormatString parses and formats a CQL string
func FormatString(input string, opts FormatOptions) (string, error) {
	return format.String(input, opts)
}

// Pretty is a convenience function for pretty formatting
func Pretty(input string) (string, error) {
	return format.PrettyString(input)
}

// Compact is a convenience function for compact formatting
func Compact(input string) (string, error) {
	return format.CompactString(input)
}

// DefaultFormatOptions returns sensible defaults for formatting
func DefaultFormatOptions() FormatOptions {
	return format.DefaultOptions()
}

// CompactFormatOptions returns options for compact formatting
func CompactFormatOptions() FormatOptions {
	return format.CompactOptions()
}

// NewSchema creates an empty schema
func NewSchema() *Schema {
	return schema.NewSchema()
}

// GetCompletions returns completion items for the given context
func GetCompletions(ctx *CompletionContext) []CompletionItem {
	return complete.GetCompletions(ctx)
}

// GetHoverInfo returns hover information for the token at the given position
func GetHoverInfo(ctx *HoverContext) *HoverInfo {
	return hover.GetHoverInfo(ctx)
}

// AnalyzeWithSchema performs full analysis of a CQL query including schema validation
// and function argument validation. This is the comprehensive analysis function.
func AnalyzeWithSchema(cql string, opts *AnalyzeOptions) *AnalyzeResult {
	return analyze.Analyze(cql, opts)
}

// DefaultAnalyzeOptions returns default analysis options
func DefaultAnalyzeOptions() *AnalyzeOptions {
	return analyze.DefaultOptions()
}

// GetTokens returns all tokens from a CQL string with semantic classification.
// The optional context provides semantic information for enhanced highlighting
// (e.g., distinguishing partition keys from regular columns).
func GetTokens(input string, ctx *TokenContext) []Token {
	return tokenize.Tokenize(input, ctx)
}
