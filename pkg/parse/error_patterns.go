package parse

import "regexp"

// ErrorPattern defines a pattern for transforming ANTLR error messages
// into user-friendly messages.
type ErrorPattern struct {
	// Name identifies this pattern (for testing/debugging)
	Name string

	// MessagePattern matches against the raw ANTLR error message
	MessagePattern *regexp.Regexp

	// QueryPattern optionally matches against the input query for context
	// If set, both MessagePattern and QueryPattern must match
	QueryPattern *regexp.Regexp

	// FriendlyTemplate is the template for the user-friendly message
	// Supports $1, $2, etc. for capture groups from MessagePattern
	FriendlyTemplate string

	// SuggestionTemplate is the template for the suggestion
	// Supports $1, $2, etc. for capture groups from MessagePattern
	SuggestionTemplate string
}

// errorPatterns is the registry of all error patterns.
// Order matters: first match wins, so more specific patterns should come first.
var errorPatterns = []ErrorPattern{
	// === Typo-specific patterns (most specific, check query content) ===

	// FORM instead of FROM
	{
		Name:               "typo-form",
		MessagePattern:     regexp.MustCompile(`(?i)no viable alternative|extraneous input`),
		QueryPattern:       regexp.MustCompile(`(?i)\bFORM\b`),
		FriendlyTemplate:   "Unknown keyword 'FORM'",
		SuggestionTemplate: "Did you mean 'FROM'?",
	},

	// WHER instead of WHERE
	{
		Name:               "typo-wher",
		MessagePattern:     regexp.MustCompile(`(?i)no viable alternative|extraneous input`),
		QueryPattern:       regexp.MustCompile(`(?i)\bWHER\b`),
		FriendlyTemplate:   "Unknown keyword 'WHER'",
		SuggestionTemplate: "Did you mean 'WHERE'?",
	},

	// SELEC/SLECT instead of SELECT
	{
		Name:               "typo-select",
		MessagePattern:     regexp.MustCompile(`(?i)mismatched input '(SELEC|SLECT|SEELCT|SELET)'`),
		FriendlyTemplate:   "Unknown keyword '$1'",
		SuggestionTemplate: "Did you mean 'SELECT'?",
	},

	// INSRT/INSER instead of INSERT
	{
		Name:               "typo-insert",
		MessagePattern:     regexp.MustCompile(`(?i)mismatched input '(INSRT|INSER|INSRET)'`),
		FriendlyTemplate:   "Unknown keyword '$1'",
		SuggestionTemplate: "Did you mean 'INSERT'?",
	},

	// UDPATE/UPDAT instead of UPDATE
	{
		Name:               "typo-update",
		MessagePattern:     regexp.MustCompile(`(?i)mismatched input '(UDPATE|UPDAT|UPDTE)'`),
		FriendlyTemplate:   "Unknown keyword '$1'",
		SuggestionTemplate: "Did you mean 'UPDATE'?",
	},

	// DELTE/DELET instead of DELETE
	{
		Name:               "typo-delete",
		MessagePattern:     regexp.MustCompile(`(?i)mismatched input '(DELTE|DELET|DELEET)'`),
		FriendlyTemplate:   "Unknown keyword '$1'",
		SuggestionTemplate: "Did you mean 'DELETE'?",
	},

	// === Generic ANTLR patterns ===

	// "no viable alternative at input 'X'" - usually means unexpected token sequence
	{
		Name:             "no-viable-alternative",
		MessagePattern:   regexp.MustCompile(`no viable alternative at input '([^']+)'`),
		FriendlyTemplate: "Unexpected syntax near '$1'",
	},

	// "mismatched input 'X' expecting {Y, Z, ...}" - wrong token
	{
		Name:               "mismatched-input-expecting",
		MessagePattern:     regexp.MustCompile(`mismatched input '([^']+)' expecting \{([^}]+)\}`),
		FriendlyTemplate:   "Unexpected '$1'",
		SuggestionTemplate: "Expected one of: $2",
	},

	// "mismatched input 'X' expecting Y" - wrong token (single expected)
	{
		Name:               "mismatched-input-single",
		MessagePattern:     regexp.MustCompile(`mismatched input '([^']+)' expecting ([^{]\S+)`),
		FriendlyTemplate:   "Unexpected '$1'",
		SuggestionTemplate: "Expected: $2",
	},

	// "extraneous input 'X' expecting Y" - extra token
	{
		Name:               "extraneous-input",
		MessagePattern:     regexp.MustCompile(`extraneous input '([^']+)' expecting`),
		FriendlyTemplate:   "Unexpected '$1'",
		SuggestionTemplate: "Remove this or check syntax",
	},

	// "missing 'X' at 'Y'" - missing required token
	{
		Name:               "missing-token",
		MessagePattern:     regexp.MustCompile(`missing '([^']+)' at '([^']+)'`),
		FriendlyTemplate:   "Missing '$1' before '$2'",
		SuggestionTemplate: "Add '$1'",
	},

	// "missing ';' at '<EOF>'" - missing semicolon
	{
		Name:               "missing-semicolon",
		MessagePattern:     regexp.MustCompile(`missing ';' at '<EOF>'`),
		FriendlyTemplate:   "Missing semicolon at end",
		SuggestionTemplate: "Add ';' at the end of the statement",
	},

	// "token recognition error at: 'X'" - invalid character
	{
		Name:               "token-recognition",
		MessagePattern:     regexp.MustCompile(`token recognition error at: '([^']+)'`),
		FriendlyTemplate:   "Invalid character '$1'",
		SuggestionTemplate: "Remove or escape this character",
	},

	// "rule X failed predicate" - semantic issue
	{
		Name:             "failed-predicate",
		MessagePattern:   regexp.MustCompile(`rule (\w+) failed predicate`),
		FriendlyTemplate: "Invalid syntax in $1",
	},
}

// GetPatterns returns all registered error patterns.
// Useful for testing.
func GetPatterns() []ErrorPattern {
	return errorPatterns
}
