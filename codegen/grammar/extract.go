// Package grammar provides functions for parsing and extracting data from CQL grammars.
package grammar

import (
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/pierre-borckmans/scql/codegen/util"
)

// ExtractANTLR4Keywords extracts K_* keywords from an ANTLR4 lexer file.
func ExtractANTLR4Keywords(path string) []string {
	content, _ := os.ReadFile(path)
	re := regexp.MustCompile(`^(K_[A-Z_0-9]+)\s*:`)

	var keywords []string
	for _, line := range strings.Split(string(content), "\n") {
		if m := re.FindStringSubmatch(line); m != nil {
			keywords = append(keywords, m[1])
		}
	}
	sort.Strings(keywords)
	return keywords
}

// ExtractScyllaKeywords extracts K_* keywords from ScyllaDB ANTLR3 grammar.
func ExtractScyllaKeywords(path string) []string {
	content, _ := os.ReadFile(path)
	re := regexp.MustCompile(`^(K_[A-Z_0-9]+):`)

	var keywords []string
	for _, line := range strings.Split(string(content), "\n") {
		if m := re.FindStringSubmatch(line); m != nil {
			keywords = append(keywords, m[1])
		}
	}
	sort.Strings(keywords)
	return keywords
}

// ExtractANTLR4Rules extracts parser rule names from an ANTLR4 parser file.
func ExtractANTLR4Rules(path string) []string {
	content, _ := os.ReadFile(path)
	// Match rule definitions like "ruleName\n    :"
	re := regexp.MustCompile(`(?m)^([a-z][a-zA-Z0-9_]*)\s*$`)

	var rules []string
	lines := strings.Split(string(content), "\n")
	for i, line := range lines {
		if m := re.FindStringSubmatch(line); m != nil {
			// Check if next non-empty line starts with ":"
			for j := i + 1; j < len(lines); j++ {
				trimmed := strings.TrimSpace(lines[j])
				if trimmed == "" {
					continue
				}
				if strings.HasPrefix(trimmed, ":") {
					rules = append(rules, m[1])
				}
				break
			}
		}
	}
	sort.Strings(rules)
	return util.Unique(rules)
}

// ExtractScyllaRules extracts parser rule names from ScyllaDB ANTLR3 grammar.
func ExtractScyllaRules(path string) []string {
	content, _ := os.ReadFile(path)
	// ScyllaDB rules have "returns" or start of rule
	re := regexp.MustCompile(`(?m)^([a-z][a-zA-Z0-9_]*)\s+(returns|\[|:)`)

	var rules []string
	for _, line := range strings.Split(string(content), "\n") {
		if m := re.FindStringSubmatch(line); m != nil {
			rules = append(rules, m[1])
		}
	}
	sort.Strings(rules)
	return util.Unique(rules)
}

// ExtractScyllaKeywordDefs extracts keyword definitions from ScyllaDB ANTLR3 grammar.
// Returns a map of keyword name to definition (e.g., "S E L E C T").
func ExtractScyllaKeywordDefs(content string) map[string]string {
	defs := make(map[string]string)
	re := regexp.MustCompile(`(?m)^(K_[A-Z_0-9]+):\s*([^;]+);`)

	for _, match := range re.FindAllStringSubmatch(content, -1) {
		if len(match) >= 3 {
			// Normalize whitespace in the definition
			def := strings.Join(strings.Fields(match[2]), " ")
			defs[match[1]] = def
		}
	}
	return defs
}
