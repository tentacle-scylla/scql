package complete

import (
	_ "embed"
	"strings"

	"github.com/tentacle-scylla/scql/gen/cqldata"
	"gopkg.in/yaml.v3"
)

//go:embed annotations.yaml
var annotationsYAML []byte

// Annotations loaded from YAML
var annotations struct {
	Keywords  map[string]annotationEntry `yaml:"keywords"`
	Functions map[string]annotationEntry `yaml:"functions"`
	Types     map[string]annotationEntry `yaml:"types"`
	Snippets  []snippetEntry             `yaml:"snippets"`
	Operators []operatorEntry            `yaml:"operators"`
}

type annotationEntry struct {
	Detail        string `yaml:"detail"`
	Documentation string `yaml:"documentation"`
	InsertText    string `yaml:"insertText"`
	Priority      int    `yaml:"priority"`
}

type snippetEntry struct {
	Label      string `yaml:"label"`
	Detail     string `yaml:"detail"`
	InsertText string `yaml:"insertText"`
	Priority   int    `yaml:"priority"`
}

type operatorEntry struct {
	Label      string `yaml:"label"`
	Detail     string `yaml:"detail"`
	InsertText string `yaml:"insertText"`
	Priority   int    `yaml:"priority"`
}

// Completion item lists - built from generated data + annotations
var (
	StatementKeywords   []CompletionItem
	ClauseKeywords      []CompletionItem
	SelectTableKeywords []CompletionItem
	WhereClauseKeywords []CompletionItem
	UpdateSetKeywords   []CompletionItem
	DeleteFromKeywords  []CompletionItem
	CreateKeywords      []CompletionItem
	AlterKeywords       []CompletionItem
	DropKeywords        []CompletionItem
	DescribeKeywords    []CompletionItem // After DESCRIBE/DESC
	PruneKeywords       []CompletionItem // After PRUNE
	CQLFunctions        []CompletionItem
	CQLTypes            []CompletionItem
	Operators           []CompletionItem
	Snippets            []CompletionItem
)

func init() {
	// Load annotations
	if err := yaml.Unmarshal(annotationsYAML, &annotations); err != nil {
		panic("failed to parse annotations.yaml: " + err.Error())
	}

	// Build completion items from generated data + annotations
	buildKeywordLists()
	buildFunctionList()
	buildTypeList()
	buildOperatorList()
	buildSnippetList()
}

func buildKeywordLists() {
	// Statement starters - per ScyllaDB docs
	starters := []string{"SELECT", "INSERT", "UPDATE", "DELETE", "CREATE", "ALTER", "DROP", "TRUNCATE", "USE", "DESCRIBE", "DESC", "BEGIN", "GRANT", "REVOKE", "LIST", "PRUNE"}
	for _, kw := range starters {
		StatementKeywords = append(StatementKeywords, makeKeywordItem(kw))
	}

	// Clause keywords (fallback)
	// Order from docs: WHERE, GROUP BY, ORDER BY, PER PARTITION LIMIT, LIMIT, ALLOW FILTERING, BYPASS CACHE, USING TIMEOUT
	clauses := []string{"WHERE", "GROUP BY", "ORDER BY", "PER PARTITION LIMIT", "LIMIT", "ALLOW FILTERING", "BYPASS CACHE", "USING TIMEOUT"}
	for _, kw := range clauses {
		ClauseKeywords = append(ClauseKeywords, makeKeywordItem(kw))
	}

	// After SELECT ... FROM table
	selectTable := []string{"WHERE", "GROUP BY", "ORDER BY", "PER PARTITION LIMIT", "LIMIT", "ALLOW FILTERING", "BYPASS CACHE", "USING TIMEOUT"}
	for _, kw := range selectTable {
		SelectTableKeywords = append(SelectTableKeywords, makeKeywordItem(kw))
	}

	// After WHERE condition
	whereClause := []string{"AND", "GROUP BY", "ORDER BY", "PER PARTITION LIMIT", "LIMIT", "ALLOW FILTERING", "BYPASS CACHE", "USING TIMEOUT"}
	for _, kw := range whereClause {
		WhereClauseKeywords = append(WhereClauseKeywords, makeKeywordItem(kw))
	}

	// After UPDATE table SET col = value
	// Note: In CQL, USING comes BEFORE SET, and IF comes AFTER WHERE
	// So after SET col = val, only WHERE is valid
	UpdateSetKeywords = []CompletionItem{
		{Label: "WHERE", Kind: KindKeyword, Detail: "Filter conditions (required)", SortPriority: 1},
	}

	// After DELETE FROM table
	// Note: IF comes after WHERE, not directly after FROM table
	DeleteFromKeywords = []CompletionItem{
		{Label: "WHERE", Kind: KindKeyword, Detail: "Filter conditions (required)", SortPriority: 1},
		{Label: "USING TIMESTAMP", Kind: KindKeyword, Detail: "Set delete timestamp", InsertText: "USING TIMESTAMP ", SortPriority: 2},
	}

	// After CREATE - per ScyllaDB docs
	createKws := []string{"TABLE", "KEYSPACE", "INDEX", "MATERIALIZED VIEW", "TYPE", "FUNCTION", "AGGREGATE", "ROLE", "USER", "TRIGGER"}
	for _, kw := range createKws {
		CreateKeywords = append(CreateKeywords, makeKeywordItem(kw))
	}

	// After ALTER - per ScyllaDB docs
	alterKws := []string{"TABLE", "KEYSPACE", "MATERIALIZED VIEW", "TYPE", "ROLE", "USER"}
	for _, kw := range alterKws {
		AlterKeywords = append(AlterKeywords, makeKeywordItem(kw))
	}

	// After DROP - per ScyllaDB docs
	dropKws := []string{"TABLE", "KEYSPACE", "INDEX", "MATERIALIZED VIEW", "TYPE", "FUNCTION", "AGGREGATE", "ROLE", "USER", "TRIGGER"}
	for _, kw := range dropKws {
		DropKeywords = append(DropKeywords, makeKeywordItem(kw))
	}

	// After DESCRIBE/DESC - per ScyllaDB docs
	// DESCRIBE can be followed by: SCHEMA, KEYSPACE, TABLE, INDEX, MATERIALIZED VIEW, TYPE, FUNCTION, AGGREGATE, CLUSTER
	describeKws := []string{"SCHEMA", "KEYSPACE", "TABLE", "INDEX", "MATERIALIZED VIEW", "TYPE", "FUNCTION", "AGGREGATE", "CLUSTER"}
	for _, kw := range describeKws {
		DescribeKeywords = append(DescribeKeywords, makeKeywordItem(kw))
	}

	// After PRUNE - ScyllaDB extension
	// PRUNE can only be followed by MATERIALIZED VIEW
	PruneKeywords = []CompletionItem{
		{Label: "MATERIALIZED VIEW", Kind: KindKeyword, Detail: "Prune ghost rows from view", InsertText: "MATERIALIZED VIEW ", SortPriority: 1},
	}
}

func buildFunctionList() {
	// Build from generated functions
	seen := make(map[string]bool)
	for _, f := range cqldata.GenFunctions {
		if seen[f.Name] {
			continue
		}
		seen[f.Name] = true

		item := CompletionItem{
			Label:        f.Name + "()",
			Kind:         KindFunction,
			InsertText:   f.Name + "()",
			SortPriority: 50,
		}

		// Add annotation details if available
		if ann, ok := annotations.Functions[f.Name]; ok {
			item.Detail = ann.Detail
			item.Documentation = ann.Documentation
			if ann.Priority > 0 {
				item.SortPriority = ann.Priority
			}
		} else {
			// Default detail from return type
			item.Detail = "Returns " + f.ReturnType
		}

		CQLFunctions = append(CQLFunctions, item)
	}
}

func buildTypeList() {
	// Build from generated types
	for _, t := range cqldata.GenTypes {
		item := CompletionItem{
			Label:        t.Name,
			Kind:         KindType,
			SortPriority: 50,
		}

		// Add annotation details if available
		if ann, ok := annotations.Types[t.Name]; ok {
			item.Detail = ann.Detail
			item.Documentation = ann.Documentation
			if ann.InsertText != "" {
				item.InsertText = ann.InsertText
			}
			if ann.Priority > 0 {
				item.SortPriority = ann.Priority
			}
		} else {
			item.Detail = t.Kind + " type"
		}

		CQLTypes = append(CQLTypes, item)
	}
}

func buildOperatorList() {
	for _, op := range annotations.Operators {
		item := CompletionItem{
			Label:        op.Label,
			Kind:         KindOperator,
			Detail:       op.Detail,
			SortPriority: op.Priority,
		}
		if op.InsertText != "" {
			item.InsertText = op.InsertText
		}
		Operators = append(Operators, item)
	}
}

func buildSnippetList() {
	for _, s := range annotations.Snippets {
		Snippets = append(Snippets, CompletionItem{
			Label:        s.Label,
			Kind:         KindSnippet,
			Detail:       s.Detail,
			InsertText:   s.InsertText,
			SortPriority: s.Priority,
		})
	}
}

func makeKeywordItem(keyword string) CompletionItem {
	item := CompletionItem{
		Label:        keyword,
		Kind:         KindKeyword,
		SortPriority: 50,
	}

	// Lookup in annotations
	if ann, ok := annotations.Keywords[keyword]; ok {
		item.Detail = ann.Detail
		if ann.InsertText != "" {
			item.InsertText = ann.InsertText
		}
		if ann.Priority > 0 {
			item.SortPriority = ann.Priority
		}
	} else {
		// Check if it's in generated keywords for validation
		upperKw := strings.ToUpper(strings.ReplaceAll(keyword, " ", "_"))
		for _, gk := range cqldata.GenAllKeywords {
			if gk == upperKw || strings.ReplaceAll(gk, "_", " ") == keyword {
				item.Detail = "CQL keyword"
				break
			}
		}
	}

	return item
}

// AllKeywords returns all CQL keywords combined
func AllKeywords() []CompletionItem {
	result := make([]CompletionItem, 0, len(StatementKeywords)+len(ClauseKeywords))
	result = append(result, StatementKeywords...)
	result = append(result, ClauseKeywords...)
	return result
}
