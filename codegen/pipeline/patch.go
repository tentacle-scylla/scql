package pipeline

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/tentacle-scylla/scql/codegen"
	"github.com/tentacle-scylla/scql/codegen/util"
)

// ParserPatch represents a patch to apply to the parser grammar.
type ParserPatch struct {
	Type    string `json:"type"`
	Rule    string `json:"rule,omitempty"`
	After   string `json:"after,omitempty"`
	Content string `json:"content"`
}

// ParserPatches represents the patches file structure.
type ParserPatches struct {
	Description string        `json:"description"`
	Version     string        `json:"version"`
	Patches     []ParserPatch `json:"patches"`
}

// Patch applies ScyllaDB extensions to the ANTLR4 grammar.
func Patch(dirs *codegen.Dirs) error {
	fmt.Println("=== Patching Grammar with ScyllaDB Extensions ===")

	// --- LEXER PATCHING ---
	fmt.Println("\n--- Patching Lexer ---")

	// Read curated keyword list from patches directory
	keywordsPath := filepath.Join(dirs.Patches, "scylla_lexer_keywords.txt")
	keywordLines := util.ReadLines(keywordsPath)

	// Parse keyword definitions (format: K_NAME:LITERAL or # comment)
	var newKeywords []string
	for _, line := range keywordLines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		kwName := strings.TrimSpace(parts[0])
		kwLiteral := strings.TrimSpace(parts[1])

		// Format: K_KEYWORD     : 'LITERAL';
		padding := 17 - len(kwName)
		if padding < 1 {
			padding = 1
		}
		newKeywords = append(newKeywords, fmt.Sprintf("%s%s: '%s';", kwName, strings.Repeat(" ", padding), kwLiteral))
		fmt.Printf("  + %s\n", kwName)
	}

	// Read original ANTLR4 lexer
	lexerContent, err := os.ReadFile(filepath.Join(dirs.Grammars, "CqlLexer.g4"))
	if err != nil {
		return fmt.Errorf("failed to read lexer: %w", err)
	}

	// Find insertion point (before "// Literals" section)
	insertPoint := strings.Index(string(lexerContent), "// Literals")
	if insertPoint == -1 {
		return fmt.Errorf("could not find insertion point in lexer")
	}

	// Insert new keywords before "// Literals"
	newLexerContent := string(lexerContent[:insertPoint]) +
		"// ScyllaDB-specific keywords\n\n" +
		strings.Join(newKeywords, "\n") + "\n\n" +
		string(lexerContent[insertPoint:])

	fmt.Printf("✓ Patched %d keywords into lexer\n", len(newKeywords))

	// Read additional lexer rules (like placeholders)
	rulesPath := filepath.Join(dirs.Patches, "scylla_lexer_rules.txt")
	ruleLines := util.ReadLines(rulesPath)

	var newRules []string
	for _, line := range ruleLines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		ruleName := strings.TrimSpace(parts[0])
		rulePattern := strings.TrimSpace(parts[1])

		// Format: RULE_NAME: pattern;
		padding := 17 - len(ruleName)
		if padding < 1 {
			padding = 1
		}
		newRules = append(newRules, fmt.Sprintf("%s%s: %s;", ruleName, strings.Repeat(" ", padding), rulePattern))
		fmt.Printf("  + %s (rule)\n", ruleName)
	}

	// Insert rules before OBJECT_NAME
	if len(newRules) > 0 {
		objNameIdx := strings.Index(newLexerContent, "OBJECT_NAME:")
		if objNameIdx != -1 {
			newLexerContent = newLexerContent[:objNameIdx] +
				"// Prepared statement placeholders\n\n" +
				strings.Join(newRules, "\n") + "\n\n" +
				newLexerContent[objNameIdx:]
			fmt.Printf("✓ Patched %d lexer rules\n", len(newRules))
		}
	}

	// Apply lexer rule replacements
	replacementsPath := filepath.Join(dirs.Patches, "scylla_lexer_replacements.txt")
	replacementLines := util.ReadLines(replacementsPath)
	for _, line := range replacementLines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		ruleName := strings.TrimSpace(parts[0])
		rulePattern := strings.TrimSpace(parts[1])

		// Find and replace the rule using regex
		ruleRegex := regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(ruleName) + `\s*:.*?;`)
		if ruleRegex.MatchString(newLexerContent) {
			newLexerContent = ruleRegex.ReplaceAllString(newLexerContent, ruleName+": "+rulePattern+";")
			fmt.Printf("  ~ %s (replaced)\n", ruleName)
		}
	}

	// Write patched lexer
	patchedLexerPath := filepath.Join(dirs.Patched, "CqlLexer.g4")
	if err := os.WriteFile(patchedLexerPath, []byte(newLexerContent), 0644); err != nil {
		return fmt.Errorf("failed to write patched lexer: %w", err)
	}

	// --- PARSER PATCHING ---
	fmt.Println("\n--- Patching Parser ---")

	// Read original parser
	parserContent, err := os.ReadFile(filepath.Join(dirs.Grammars, "CqlParser.g4"))
	if err != nil {
		return fmt.Errorf("failed to read parser: %w", err)
	}
	parserStr := string(parserContent)

	// Try to load parser patches file
	patchesPath := filepath.Join(dirs.Patches, "scylla_parser_patches.json")
	patchesData, err := os.ReadFile(patchesPath)
	if err != nil {
		fmt.Printf("No parser patches file found at %s, copying parser unchanged\n", patchesPath)
	} else {
		var patches ParserPatches
		if err := json.Unmarshal(patchesData, &patches); err != nil {
			return fmt.Errorf("failed to parse patches file: %w", err)
		}
		fmt.Printf("Applying %d parser patches (ScyllaDB %s)\n", len(patches.Patches), patches.Version)
		parserStr = applyParserPatches(parserStr, patches.Patches)
	}

	// Write patched parser
	patchedParserPath := filepath.Join(dirs.Patched, "CqlParser.g4")
	if err := os.WriteFile(patchedParserPath, []byte(parserStr), 0644); err != nil {
		return fmt.Errorf("failed to write patched parser: %w", err)
	}
	fmt.Printf("✓ Parser patched and written\n")

	return nil
}

// applyParserPatches applies a list of patches to parser content.
func applyParserPatches(content string, patches []ParserPatch) string {
	for _, patch := range patches {
		switch patch.Type {
		case "add_to_cql_rule":
			content = addToCqlRule(content, patch.Content)
		case "replace_rule":
			content = replaceRule(content, patch.Rule, patch.Content)
		case "add_rule":
			content = addRuleAfter(content, patch.After, patch.Content)
		case "add_to_rule":
			content = addToRule(content, patch.Rule, patch.Content)
		case "add_keywords":
			content = addKeywordRules(content, patch.Content)
		default:
			fmt.Printf("  ? Unknown patch type: %s\n", patch.Type)
		}
	}
	return content
}

// addToCqlRule adds an alternative to the cql rule.
func addToCqlRule(content, alternative string) string {
	re := regexp.MustCompile(`(?s)(cql\s*:\s*[^;]+)(;)`)
	return re.ReplaceAllString(content, "${1}\n    "+alternative+"\n    ${2}")
}

// replaceRule replaces an entire rule definition.
func replaceRule(content, ruleName, newRule string) string {
	pattern := fmt.Sprintf(`(?s)%s\s*:\s*[^;]+;`, regexp.QuoteMeta(ruleName))
	re := regexp.MustCompile(pattern)
	return re.ReplaceAllString(content, newRule)
}

// addRuleAfter adds a new rule after an existing rule.
func addRuleAfter(content, afterRule, newRule string) string {
	pattern := fmt.Sprintf(`(?s)(%s\s*:\s*[^;]+;)`, regexp.QuoteMeta(afterRule))
	re := regexp.MustCompile(pattern)
	return re.ReplaceAllString(content, "${1}\n\n"+newRule)
}

// addToRule adds alternatives to an existing rule (before the semicolon).
func addToRule(content, ruleName, alternatives string) string {
	pattern := fmt.Sprintf(`(?s)(%s\s*:\s*[^;]+)(;)`, regexp.QuoteMeta(ruleName))
	re := regexp.MustCompile(pattern)
	return re.ReplaceAllString(content, "${1}\n    "+alternatives+"\n    ${2}")
}

// addKeywordRules adds keyword wrapper rules before the final part of the grammar.
func addKeywordRules(content, rules string) string {
	if idx := strings.LastIndex(content, "fragment "); idx != -1 {
		return content[:idx] + rules + "\n\n" + content[idx:]
	}
	return content + "\n" + rules + "\n"
}
