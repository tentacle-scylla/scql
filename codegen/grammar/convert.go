// Package grammar provides functions for parsing and extracting data from CQL grammars.
package grammar

import (
	"fmt"
	"regexp"
	"strings"
)

// ExtractScyllaParserRules extracts parser rules from ScyllaDB ANTLR3 grammar.
// Returns map of ruleName -> rule body (with C++ stripped).
func ExtractScyllaParserRules(content string) map[string]string {
	rules := make(map[string]string)

	// Split by rule definitions (lowercase identifier at start of line followed by returns or :)
	ruleStartRe := regexp.MustCompile(`(?m)^([a-z][a-zA-Z0-9_]*)\s*(returns\s*\[[^\]]+\])?\s*$`)

	lines := strings.Split(content, "\n")
	var currentRule string
	var ruleBody strings.Builder
	inRule := false
	braceDepth := 0

	for i := 0; i < len(lines); i++ {
		line := lines[i]

		// Check for rule start
		if m := ruleStartRe.FindStringSubmatch(line); m != nil {
			// Save previous rule if any
			if currentRule != "" && ruleBody.Len() > 0 {
				rules[currentRule] = ConvertRuleToANTLR4(ruleBody.String())
			}
			currentRule = m[1]
			ruleBody.Reset()
			inRule = false
			braceDepth = 0
			continue
		}

		// Skip @init blocks
		if strings.Contains(line, "@init") {
			// Skip until closing brace
			for i < len(lines) && !strings.Contains(lines[i], "}") {
				i++
			}
			continue
		}

		// Look for rule body start (:)
		if currentRule != "" && !inRule {
			if idx := strings.Index(line, ":"); idx >= 0 {
				inRule = true
				// Add everything after the colon
				ruleBody.WriteString(line[idx:])
				ruleBody.WriteString("\n")
			}
			continue
		}

		// In rule body - collect until semicolon at depth 0
		if inRule {
			ruleBody.WriteString(line)
			ruleBody.WriteString("\n")

			// Track brace depth for action blocks
			braceDepth += strings.Count(line, "{") - strings.Count(line, "}")

			// Check for rule end (semicolon not inside braces)
			if braceDepth == 0 && strings.Contains(line, ";") {
				rules[currentRule] = ConvertRuleToANTLR4(ruleBody.String())
				currentRule = ""
				ruleBody.Reset()
				inRule = false
			}
		}
	}

	return rules
}

// ConvertRuleToANTLR4 converts an ANTLR3 rule body to ANTLR4 format
// by stripping C++ actions, variable assignments, and returns clauses.
func ConvertRuleToANTLR4(body string) string {
	// Remove { ... } action blocks (handling nested braces)
	body = RemoveActionBlocks(body)

	// Remove variable assignments (name= before rule/token references)
	varAssignRe := regexp.MustCompile(`[a-z_][a-z0-9_]*\s*=\s*`)
	body = varAssignRe.ReplaceAllString(body, "")

	// Remove $var references
	dollarVarRe := regexp.MustCompile(`\$[a-z_][a-z0-9_]*`)
	body = dollarVarRe.ReplaceAllString(body, "")

	// Clean up extra whitespace
	body = strings.Join(strings.Fields(body), " ")

	// Ensure proper formatting
	body = strings.ReplaceAll(body, " ;", ";")
	body = strings.ReplaceAll(body, " )", ")")
	body = strings.ReplaceAll(body, "( ", "(")

	return body
}

// RemoveActionBlocks removes { ... } blocks handling nested braces.
func RemoveActionBlocks(s string) string {
	var result strings.Builder
	depth := 0
	i := 0

	for i < len(s) {
		if s[i] == '{' {
			depth++
			i++
			continue
		}
		if s[i] == '}' {
			depth--
			i++
			continue
		}
		if depth == 0 {
			result.WriteByte(s[i])
		}
		i++
	}

	return result.String()
}

// ConvertToANTLR4Keyword converts ScyllaDB ANTLR3 keyword def to ANTLR4 format.
// Input:  "S E L E C T" or "S E R V I C E '_' L E V E L"
// Output: "K_SELECT : 'SELECT';" or "K_SERVICE_LEVEL : 'SERVICE_LEVEL';"
func ConvertToANTLR4Keyword(keyword string, antlr3Def string) string {
	// Handle multi-word keywords like K_TABLES with alternatives
	if strings.Contains(antlr3Def, "(") {
		// Complex definition with alternatives - extract first option
		// e.g., "( C O L U M N F A M I L I E S | T A B L E S )"
		re := regexp.MustCompile(`\(\s*([^|)]+)`)
		if m := re.FindStringSubmatch(antlr3Def); m != nil {
			antlr3Def = strings.TrimSpace(m[1])
		}
	}

	// Convert space-separated letters to word
	// "S E L E C T" -> "SELECT"
	// "S E R V I C E '_' L E V E L" -> "SERVICE_LEVEL"
	var result strings.Builder
	parts := strings.Fields(antlr3Def)
	for _, p := range parts {
		if p == "'_'" {
			result.WriteRune('_')
		} else if len(p) == 1 && p[0] >= 'A' && p[0] <= 'Z' {
			result.WriteString(p)
		}
	}

	word := result.String()
	if word == "" {
		// Fallback: derive from keyword name
		word = strings.TrimPrefix(keyword, "K_")
	}

	// Calculate padding for alignment
	padding := 17 - len(keyword)
	if padding < 1 {
		padding = 1
	}

	return fmt.Sprintf("%s%s: '%s';", keyword, strings.Repeat(" ", padding), word)
}
