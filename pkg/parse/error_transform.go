package parse

import (
	"regexp"
	"strings"
)

// TransformResult contains the transformed error message and suggestion
type TransformResult struct {
	FriendlyMessage string
	Suggestion      string
	PatternName     string // Which pattern matched (for debugging)
}

// TransformError transforms a raw ANTLR error message into a user-friendly message.
// It searches through the pattern registry for a match.
// If no pattern matches, it returns the raw message as-is.
func TransformError(rawMessage string, query string) TransformResult {
	for _, pattern := range errorPatterns {
		matches := pattern.MessagePattern.FindStringSubmatch(rawMessage)
		if matches == nil {
			continue
		}

		// If query pattern is specified, it must also match
		if pattern.QueryPattern != nil && !pattern.QueryPattern.MatchString(query) {
			continue
		}

		// Pattern matched - expand templates
		friendly := expandTemplate(pattern.FriendlyTemplate, matches)
		suggestion := expandTemplate(pattern.SuggestionTemplate, matches)

		// If no friendly template, fall back to raw message
		if friendly == "" {
			friendly = rawMessage
		}

		// Try to find a typo suggestion - this should override generic "expected" suggestions
		typoSuggestion := findTypoSuggestion(rawMessage, query)
		if typoSuggestion != "" {
			// If we found a typo, update the friendly message too
			token := extractToken(rawMessage)
			if token != "" && SuggestKeyword(token) != "" {
				friendly = "Unknown keyword '" + token + "'"
			}
			suggestion = typoSuggestion
		}

		return TransformResult{
			FriendlyMessage: friendly,
			Suggestion:      suggestion,
			PatternName:     pattern.Name,
		}
	}

	// No pattern matched - return raw message
	return TransformResult{
		FriendlyMessage: rawMessage,
		Suggestion:      "",
		PatternName:     "",
	}
}

// findTypoSuggestion extracts tokens from the error message and query,
// and returns a typo suggestion if one is found.
func findTypoSuggestion(rawMessage string, query string) string {
	// Try to extract the mismatched token from the error message
	patterns := []string{
		`mismatched input '([^']+)'`,
		`extraneous input '([^']+)'`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(rawMessage); len(matches) > 1 {
			token := matches[1]
			if suggestion := SuggestKeyword(token); suggestion != "" {
				return "Did you mean '" + suggestion + "'?"
			}
		}
	}

	// For "no viable alternative at input 'X Y Z'" - try the last word
	noViableRe := regexp.MustCompile(`no viable alternative at input '([^']+)'`)
	if matches := noViableRe.FindStringSubmatch(rawMessage); len(matches) > 1 {
		words := strings.Fields(matches[1])
		if len(words) > 0 {
			lastWord := words[len(words)-1]
			lastWord = strings.TrimRight(lastWord, ";,()[]{}=<>!@#$%^&*")
			if suggestion := SuggestKeyword(lastWord); suggestion != "" {
				return "Did you mean '" + suggestion + "'?"
			}
		}
	}

	return ""
}

// extractToken extracts the problematic token from an error message.
func extractToken(rawMessage string) string {
	patterns := []string{
		`mismatched input '([^']+)'`,
		`extraneous input '([^']+)'`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(rawMessage); len(matches) > 1 {
			return matches[1]
		}
	}
	return ""
}

// expandTemplate expands a template string with capture group values.
// $1, $2, etc. are replaced with the corresponding matches.
func expandTemplate(template string, matches []string) string {
	if template == "" {
		return ""
	}

	result := template
	for i := 1; i < len(matches); i++ {
		placeholder := "$" + string(rune('0'+i))
		result = strings.ReplaceAll(result, placeholder, matches[i])
	}
	return result
}

// CleanExpectingList cleans up ANTLR's verbose "expecting {X, Y, Z}" lists
// by removing quotes and limiting the number of items shown.
func CleanExpectingList(raw string, maxItems int) string {
	// Remove surrounding braces if present
	raw = strings.TrimPrefix(raw, "{")
	raw = strings.TrimSuffix(raw, "}")

	// Split by comma
	items := strings.Split(raw, ",")

	// Clean each item
	cleaned := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		item = strings.Trim(item, "'\"")
		if item != "" && item != "<EOF>" {
			cleaned = append(cleaned, item)
		}
	}

	// Limit items
	if len(cleaned) > maxItems {
		cleaned = append(cleaned[:maxItems], "...")
	}

	return strings.Join(cleaned, ", ")
}
