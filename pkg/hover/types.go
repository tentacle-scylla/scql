// Package hover provides type information for CQL tokens at cursor positions.
package hover

import "github.com/tentacle-scylla/scql/pkg/schema"

// HoverKind identifies the type of hover target.
type HoverKind string

const (
	HoverKeyword  HoverKind = "keyword"
	HoverFunction HoverKind = "function"
	HoverTable    HoverKind = "table"
	HoverColumn   HoverKind = "column"
	HoverKeyspace HoverKind = "keyspace"
	HoverType     HoverKind = "type"
	HoverOperator HoverKind = "operator"
)

// Range represents a text range in the query.
type Range struct {
	// Start is the starting offset (inclusive)
	Start int `json:"start"`
	// End is the ending offset (exclusive)
	End int `json:"end"`
}

// HoverInfo contains information about a token at a position.
type HoverInfo struct {
	// Content is the hover text (supports markdown)
	Content string `json:"content"`

	// Range is the text range this hover applies to
	Range *Range `json:"range,omitempty"`

	// Kind identifies the type of token
	Kind HoverKind `json:"kind"`

	// Name is the token name (e.g., column name, function name)
	Name string `json:"name"`
}

// HoverContext provides context for hover resolution.
type HoverContext struct {
	// Query is the full CQL query text
	Query string

	// Position is the cursor offset in the query
	Position int

	// Schema is the optional schema for context-aware hovers
	Schema *schema.Schema

	// DefaultKeyspace is used when no keyspace is specified
	DefaultKeyspace string
}

// Token represents a lexical token found at a position.
type Token struct {
	// Text is the token text
	Text string

	// Start is the starting offset
	Start int

	// End is the ending offset
	End int

	// Type indicates the token type
	Type TokenType
}

// TokenType identifies the type of token.
type TokenType string

const (
	TokenUnknown     TokenType = "unknown"
	TokenIdentifier  TokenType = "identifier"
	TokenKeyword     TokenType = "keyword"
	TokenFunction    TokenType = "function"
	TokenOperator    TokenType = "operator"
	TokenLiteral     TokenType = "literal"
	TokenPunctuation TokenType = "punctuation"
)
