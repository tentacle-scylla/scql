// Package complete provides context-aware CQL auto-completion.
package complete

import "github.com/tentacle-scylla/scql/pkg/schema"

// CompletionKind identifies the type of completion item.
type CompletionKind string

const (
	KindKeyword  CompletionKind = "keyword"
	KindTable    CompletionKind = "table"
	KindView     CompletionKind = "view" // Materialized view
	KindColumn   CompletionKind = "column"
	KindFunction CompletionKind = "function"
	KindType     CompletionKind = "type"
	KindKeyspace CompletionKind = "keyspace"
	KindOperator CompletionKind = "operator"
	KindSnippet  CompletionKind = "snippet"
)

// GroupKind identifies what type of grouping this is.
type GroupKind string

const (
	GroupKindKeyspace GroupKind = "keyspace" // Tables/views by keyspace
	GroupKindCategory GroupKind = "category" // Functions/types by category
	GroupKindKeyType  GroupKind = "key_type" // Columns by partition/clustering/regular
	GroupKindSource   GroupKind = "source"   // Columns by source table
)

// CompletionGroup represents a group that completion items can belong to.
// Items can belong to multiple groups (e.g., a column can be in both
// "source:users" and "key:partition" groups).
type CompletionGroup struct {
	// ID is a unique identifier for the group (e.g., "ks:myapp", "cat:aggregate", "key:partition")
	ID string `json:"id"`

	// Kind identifies what type of grouping this is
	Kind GroupKind `json:"kind"`

	// Label is the display text for the group (e.g., "myapp", "Aggregate", "Partition Key")
	Label string `json:"label"`

	// Icon is an optional icon hint for the frontend (e.g., "⬡", "ƒ", "★")
	Icon string `json:"icon,omitempty"`

	// Priority controls ordering of groups (lower = first)
	Priority int `json:"priority,omitempty"`
}

// CompletionItem represents a single completion suggestion.
type CompletionItem struct {
	// Label is the display text shown in the completion list
	Label string `json:"label"`

	// Kind identifies the type of completion (keyword, table, column, etc.)
	Kind CompletionKind `json:"kind"`

	// Detail provides additional info (e.g., column type, function signature)
	Detail string `json:"detail,omitempty"`

	// InsertText is the text to insert (may differ from Label for snippets)
	InsertText string `json:"insertText,omitempty"`

	// Documentation provides extended documentation
	Documentation string `json:"documentation,omitempty"`

	// SortPriority controls ordering (lower = higher priority)
	SortPriority int `json:"sortPriority,omitempty"`

	// FilterText is used for filtering (defaults to Label if empty)
	FilterText string `json:"filterText,omitempty"`

	// Groups contains IDs of groups this item belongs to.
	// An item can belong to multiple groups (e.g., ["ks:myapp", "key:partition"]).
	Groups []string `json:"groups,omitempty"`
}

// CompletionResult contains completions along with group definitions.
type CompletionResult struct {
	// Groups contains all group definitions referenced by items.
	// The frontend can use these to render grouped views or badges.
	Groups []CompletionGroup `json:"groups"`

	// Items contains the completion suggestions.
	Items []CompletionItem `json:"items"`

	// Context indicates what kind of completion context we're in.
	// Frontend can use this to determine UI behavior (e.g., multi-select in SELECT).
	Context ContextType `json:"context"`

	// AllowMultiSelect indicates whether multiple items can be selected.
	// True for SELECT column lists, INSERT column lists.
	AllowMultiSelect bool `json:"allowMultiSelect"`

	// SelectedColumns contains columns already present in the SELECT clause.
	// Frontend can use this to pre-check checkboxes in multi-select mode.
	SelectedColumns []string `json:"selectedColumns,omitempty"`
}

// GetInsertText returns the text to insert, defaulting to Label.
func (c *CompletionItem) GetInsertText() string {
	if c.InsertText != "" {
		return c.InsertText
	}
	return c.Label
}

// ContextType identifies what kind of completion context we're in.
type ContextType string

const (
	ContextUnknown          ContextType = "unknown"
	ContextStatementStart   ContextType = "statement_start"   // Beginning of a statement
	ContextAfterSelect        ContextType = "after_select"         // After SELECT keyword
	ContextAfterSelectColumns ContextType = "after_select_columns" // After SELECT columns, need FROM
	ContextAfterFrom          ContextType = "after_from"           // After FROM keyword
	ContextAfterSelectTable   ContextType = "after_select_table"   // After SELECT ... FROM table
	ContextAfterWhere       ContextType = "after_where"       // After WHERE keyword
	ContextAfterAnd         ContextType = "after_and"         // After AND/OR in WHERE
	ContextAfterOperator    ContextType = "after_operator"    // After =, <, >, etc.
	ContextAfterInsertInto  ContextType = "after_insert_into" // After INSERT INTO
	ContextAfterUpdate      ContextType = "after_update"      // After UPDATE keyword
	ContextAfterSet         ContextType = "after_set"         // After SET in UPDATE
	ContextAfterDelete      ContextType = "after_delete"      // After DELETE keyword
	ContextAfterOrderBy     ContextType = "after_order_by"    // After ORDER BY
	ContextAfterGroupBy     ContextType = "after_group_by"    // After GROUP BY
	ContextAfterCreate      ContextType = "after_create"      // After CREATE keyword
	ContextAfterAlter       ContextType = "after_alter"       // After ALTER keyword
	ContextAfterDrop        ContextType = "after_drop"        // After DROP keyword
	ContextAfterUse         ContextType = "after_use"         // After USE keyword
	ContextAfterDot         ContextType = "after_dot"         // After keyspace.
	ContextInColumnList     ContextType = "in_column_list"    // Inside (col1, col2, ...)
	ContextAfterValues      ContextType = "after_values"      // After VALUES keyword
	ContextInTypeSpec       ContextType = "in_type_spec"      // Inside type specification
	ContextAfterColumn      ContextType = "after_column"      // After column name in WHERE
	ContextAfterUpdateSet   ContextType = "after_update_set"  // After UPDATE table SET col = value
	ContextAfterDeleteFrom  ContextType = "after_delete_from" // After DELETE FROM table
	ContextAfterLimit       ContextType = "after_limit"       // After LIMIT keyword (expects number)
	ContextAfterLimitValue  ContextType = "after_limit_value" // After LIMIT <number> (only ALLOW FILTERING valid)
	ContextAfterDescribe    ContextType = "after_describe"    // After DESCRIBE/DESC keyword
	ContextAfterPrune       ContextType = "after_prune"       // After PRUNE keyword
)

// CompletionContext contains all information needed to generate completions.
type CompletionContext struct {
	// Query is the full CQL query text
	Query string

	// Position is the cursor offset in the query
	Position int

	// Schema is the optional schema for context-aware completions
	Schema *schema.Schema

	// DefaultKeyspace is used when no keyspace is specified
	DefaultKeyspace string
}

// DetectedContext contains the result of analyzing the cursor position.
type DetectedContext struct {
	// Type is the detected context type
	Type ContextType

	// Prefix is the text being typed (for filtering)
	Prefix string

	// Keyspace is the current keyspace context (if any)
	Keyspace string

	// Table is the current table context (if any)
	Table string

	// Column is the current column context (e.g., in WHERE col = |)
	Column string

	// TokenStart is the start position of the current token
	TokenStart int

	// TokenEnd is the end position of the current token
	TokenEnd int
}

// CompletionOptions configures completion behavior.
type CompletionOptions struct {
	// MaxItems limits the number of returned completions (0 = unlimited)
	MaxItems int

	// IncludeSnippets includes template snippets in results
	IncludeSnippets bool

	// IncludeKeywords includes CQL keywords in results
	IncludeKeywords bool

	// IncludeFunctions includes built-in functions in results
	IncludeFunctions bool

	// IncludeTypes includes CQL types in results
	IncludeTypes bool

	// UseANTLRFilter enables ANTLR-based validation to filter out
	// grammatically invalid keyword suggestions. This provides more
	// accurate completions but has a small performance cost.
	UseANTLRFilter bool
}

// DefaultOptions returns the default completion options.
func DefaultOptions() *CompletionOptions {
	return &CompletionOptions{
		MaxItems:         50,
		IncludeSnippets:  true,
		IncludeKeywords:  true,
		IncludeFunctions: true,
		IncludeTypes:     true,
		UseANTLRFilter:   true,
	}
}
