package cqlextract

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
)

// GrammarData holds extracted keyword information from the ANTLR3 grammar.
type GrammarData struct {
	AllKeywords        []string // All K_* tokens
	UnreservedKeywords []string // Keywords that can be used as identifiers
	BasicUnreserved    []string // Basic unreserved keywords
	TypeKeywords       []string // Type keywords (ASCII, BIGINT, etc.)

	// Context-based keyword groupings (extracted from grammar rules)
	AfterCreate []string // Keywords valid after CREATE
	AfterAlter  []string // Keywords valid after ALTER
	AfterDrop   []string // Keywords valid after DROP
}

// ParseGrammar extracts keywords from an ANTLR3 grammar file.
func ParseGrammar(grammarPath string) (*GrammarData, error) {
	content, err := os.ReadFile(grammarPath)
	if err != nil {
		return nil, fmt.Errorf("reading grammar file: %w", err)
	}

	text := string(content)
	data := &GrammarData{}

	// Extract all K_* keyword definitions
	data.AllKeywords = extractKeywordTokens(text)

	// Extract keyword categories from grammar rules
	data.BasicUnreserved = extractRuleKeywords(text, "basic_unreserved_keyword")
	data.TypeKeywords = extractRuleKeywords(text, "type_unreserved_keyword")

	// unreserved_keyword = basic_unreserved + type_unreserved + (TTL, COUNT, WRITETIME, KEY)
	data.UnreservedKeywords = append(data.UnreservedKeywords, data.BasicUnreserved...)
	data.UnreservedKeywords = append(data.UnreservedKeywords, data.TypeKeywords...)
	data.UnreservedKeywords = append(data.UnreservedKeywords, "TTL", "COUNT", "WRITETIME", "KEY")

	// Extract context-based keywords from grammar rules
	data.AfterCreate = extractContextKeywords(text, "createStatement")
	data.AfterAlter = extractContextKeywords(text, "alterStatement")
	data.AfterDrop = extractContextKeywords(text, "dropStatement")

	return data, nil
}

// extractKeywordTokens finds all K_KEYWORD token definitions.
func extractKeywordTokens(text string) []string {
	re := regexp.MustCompile(`(?m)^K_([A-Z_]+)\s*:\s*[^;]+;`)

	matches := re.FindAllStringSubmatch(text, -1)
	keywords := make([]string, 0, len(matches))
	seen := make(map[string]bool)

	for _, match := range matches {
		if len(match) > 1 {
			keyword := match[1]
			if !seen[keyword] {
				seen[keyword] = true
				keywords = append(keywords, keyword)
			}
		}
	}

	sort.Strings(keywords)
	return keywords
}

// extractRuleKeywords extracts keywords from a grammar rule like basic_unreserved_keyword.
func extractRuleKeywords(text string, ruleName string) []string {
	rulePattern := regexp.MustCompile(fmt.Sprintf(`(?s)%s\s+returns\s+\[[^\]]+\]\s*:\s*k=\(\s*([^)]+)\)`, ruleName))
	match := rulePattern.FindStringSubmatch(text)
	if match == nil || len(match) < 2 {
		return nil
	}

	keywordPattern := regexp.MustCompile(`K_([A-Z_]+)`)
	keywordMatches := keywordPattern.FindAllStringSubmatch(match[1], -1)

	keywords := make([]string, 0, len(keywordMatches))
	for _, km := range keywordMatches {
		if len(km) > 1 {
			keywords = append(keywords, km[1])
		}
	}

	sort.Strings(keywords)
	return keywords
}

// extractContextKeywords extracts keywords that follow a specific statement type.
func extractContextKeywords(text string, statementType string) []string {
	switch statementType {
	case "createStatement":
		return extractCreateContextKeywords(text)
	case "alterStatement":
		return extractAlterContextKeywords(text)
	case "dropStatement":
		return extractDropContextKeywords(text)
	}
	return nil
}

// extractCreateContextKeywords extracts keywords valid after CREATE.
func extractCreateContextKeywords(text string) []string {
	keywords := []string{}

	patterns := []string{
		`createKeyspaceStatement`,
		`createTableStatement`,
		`createIndexStatement`,
		`createViewStatement`,
		`createTypeStatement`,
		`createFunctionStatement`,
		`createAggregateStatement`,
		`createRoleStatement`,
		`createUserStatement`,
		`createServiceLevelStatement`,
		`createTriggerStatement`,
	}

	for _, p := range patterns {
		if strings.Contains(text, p) {
			name := strings.TrimPrefix(p, "create")
			name = strings.TrimSuffix(name, "Statement")
			name = strings.ToUpper(name)

			switch name {
			case "KEYSPACE":
				keywords = append(keywords, "KEYSPACE")
			case "TABLE":
				keywords = append(keywords, "TABLE")
			case "INDEX":
				keywords = append(keywords, "INDEX")
			case "VIEW":
				keywords = append(keywords, "MATERIALIZED VIEW")
			case "TYPE":
				keywords = append(keywords, "TYPE")
			case "FUNCTION":
				keywords = append(keywords, "FUNCTION")
			case "AGGREGATE":
				keywords = append(keywords, "AGGREGATE")
			case "ROLE":
				keywords = append(keywords, "ROLE")
			case "USER":
				keywords = append(keywords, "USER")
			case "SERVICELEVEL":
				keywords = append(keywords, "SERVICE LEVEL")
			case "TRIGGER":
				keywords = append(keywords, "TRIGGER")
			}
		}
	}

	return keywords
}

// extractAlterContextKeywords extracts keywords valid after ALTER.
func extractAlterContextKeywords(text string) []string {
	keywords := []string{}

	patterns := []string{
		`alterKeyspaceStatement`,
		`alterTableStatement`,
		`alterTypeStatement`,
		`alterRoleStatement`,
		`alterUserStatement`,
		`alterServiceLevelStatement`,
	}

	for _, p := range patterns {
		if strings.Contains(text, p) {
			name := strings.TrimPrefix(p, "alter")
			name = strings.TrimSuffix(name, "Statement")
			name = strings.ToUpper(name)

			switch name {
			case "KEYSPACE":
				keywords = append(keywords, "KEYSPACE")
			case "TABLE":
				keywords = append(keywords, "TABLE")
			case "TYPE":
				keywords = append(keywords, "TYPE")
			case "ROLE":
				keywords = append(keywords, "ROLE")
			case "USER":
				keywords = append(keywords, "USER")
			case "SERVICELEVEL":
				keywords = append(keywords, "SERVICE LEVEL")
			}
		}
	}

	return keywords
}

// extractDropContextKeywords extracts keywords valid after DROP.
func extractDropContextKeywords(text string) []string {
	keywords := []string{}

	patterns := []string{
		`dropKeyspaceStatement`,
		`dropTableStatement`,
		`dropIndexStatement`,
		`dropViewStatement`,
		`dropTypeStatement`,
		`dropFunctionStatement`,
		`dropAggregateStatement`,
		`dropRoleStatement`,
		`dropUserStatement`,
		`dropServiceLevelStatement`,
		`dropTriggerStatement`,
	}

	for _, p := range patterns {
		if strings.Contains(text, p) {
			name := strings.TrimPrefix(p, "drop")
			name = strings.TrimSuffix(name, "Statement")
			name = strings.ToUpper(name)

			switch name {
			case "KEYSPACE":
				keywords = append(keywords, "KEYSPACE")
			case "TABLE":
				keywords = append(keywords, "TABLE")
			case "INDEX":
				keywords = append(keywords, "INDEX")
			case "VIEW":
				keywords = append(keywords, "MATERIALIZED VIEW")
			case "TYPE":
				keywords = append(keywords, "TYPE")
			case "FUNCTION":
				keywords = append(keywords, "FUNCTION")
			case "AGGREGATE":
				keywords = append(keywords, "AGGREGATE")
			case "ROLE":
				keywords = append(keywords, "ROLE")
			case "USER":
				keywords = append(keywords, "USER")
			case "SERVICELEVEL":
				keywords = append(keywords, "SERVICE LEVEL")
			case "TRIGGER":
				keywords = append(keywords, "TRIGGER")
			}
		}
	}

	return keywords
}

// CategorizeKeywords splits keywords into statement starters and clause keywords.
func CategorizeKeywords(data *GrammarData) (starters []string, clauses []string) {
	starterSet := map[string]bool{
		"SELECT": true, "INSERT": true, "UPDATE": true, "DELETE": true,
		"CREATE": true, "ALTER": true, "DROP": true, "TRUNCATE": true,
		"USE": true, "DESCRIBE": true, "DESC": true, "BEGIN": true,
		"GRANT": true, "REVOKE": true, "LIST": true, "BATCH": true,
		"APPLY": true,
	}

	clauseSet := map[string]bool{
		"FROM": true, "WHERE": true, "AND": true, "OR": true,
		"ORDER": true, "BY": true, "GROUP": true, "LIMIT": true,
		"ALLOW": true, "FILTERING": true, "USING": true,
		"SET": true, "VALUES": true, "INTO": true, "IF": true,
		"EXISTS": true, "NOT": true, "IN": true, "CONTAINS": true,
		"KEY": true, "PRIMARY": true, "WITH": true, "ASC": true,
		"DESC": true, "DISTINCT": true, "AS": true, "TTL": true,
		"TIMESTAMP": true, "JSON": true, "TOKEN": true,
		"WRITETIME": true, "PARTITION": true, "PER": true,
		"STATIC": true, "FROZEN": true, "CLUSTERING": true,
	}

	for _, kw := range data.AllKeywords {
		if starterSet[kw] {
			starters = append(starters, kw)
		}
		if clauseSet[kw] {
			clauses = append(clauses, kw)
		}
	}

	sort.Strings(starters)
	sort.Strings(clauses)
	return starters, clauses
}

// ParseGrammarFromReader parses grammar from an io.Reader (for testing).
func ParseGrammarFromReader(r *bufio.Reader) (*GrammarData, error) {
	var builder strings.Builder
	for {
		line, err := r.ReadString('\n')
		builder.WriteString(line)
		if err != nil {
			break
		}
	}

	text := builder.String()
	data := &GrammarData{}

	data.AllKeywords = extractKeywordTokens(text)
	data.BasicUnreserved = extractRuleKeywords(text, "basic_unreserved_keyword")
	data.TypeKeywords = extractRuleKeywords(text, "type_unreserved_keyword")

	data.UnreservedKeywords = append(data.UnreservedKeywords, data.BasicUnreserved...)
	data.UnreservedKeywords = append(data.UnreservedKeywords, data.TypeKeywords...)
	data.UnreservedKeywords = append(data.UnreservedKeywords, "TTL", "COUNT", "WRITETIME", "KEY")

	data.AfterCreate = extractCreateContextKeywords(text)
	data.AfterAlter = extractAlterContextKeywords(text)
	data.AfterDrop = extractDropContextKeywords(text)

	return data, nil
}
