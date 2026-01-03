package complete

import (
	"strings"

	"github.com/antlr4-go/antlr/v4"
	parser "github.com/tentacle-scylla/scql/gen/parser"
)

// PredictedTokens contains the result of ANTLR-based token prediction.
type PredictedTokens struct {
	// ExpectedTokenTypes contains token type IDs that are valid at the cursor position
	ExpectedTokenTypes []int
	// TokenNames maps token type IDs to their symbolic names
	TokenNames map[int]string
}

// predictingErrorListener captures expected tokens when syntax errors occur.
type predictingErrorListener struct {
	*antlr.DefaultErrorListener
	expectedTokens *antlr.IntervalSet
	symbolicNames  []string
	literalNames   []string
	captured       bool
	parser         antlr.Parser // Store parser reference for later use
}

func (p *predictingErrorListener) SyntaxError(
	recognizer antlr.Recognizer,
	offendingSymbol any,
	line, column int,
	msg string,
	e antlr.RecognitionException,
) {
	// Only capture from the first error (closest to cursor position)
	if p.captured {
		return
	}

	// Try to get expected tokens from the parser
	if pr, ok := recognizer.(antlr.Parser); ok {
		p.parser = pr
		// Use defer/recover since GetExpectedTokens can panic in some states
		func() {
			defer func() {
				_ = recover() // Silently recover from any panic
			}()
			p.expectedTokens = pr.GetExpectedTokens()
			p.symbolicNames = pr.GetSymbolicNames()
			p.literalNames = pr.GetLiteralNames()
			p.captured = true
		}()
	}
}

// captureFromParser attempts to get expected tokens directly from parser after parsing
func (p *predictingErrorListener) captureFromParser(pr antlr.Parser) {
	if p.captured {
		return
	}
	p.parser = pr
	func() {
		defer func() {
			_ = recover()
		}()
		p.expectedTokens = pr.GetExpectedTokens()
		p.symbolicNames = pr.GetSymbolicNames()
		p.literalNames = pr.GetLiteralNames()
		p.captured = true
	}()
}

// GetExpectedTokensAtPosition parses the query up to the cursor position
// and returns the set of tokens that would be valid at that position.
func GetExpectedTokensAtPosition(query string, position int) *PredictedTokens {
	// Truncate query at cursor position
	if position > len(query) {
		position = len(query)
	}
	truncated := query[:position]

	// Try multiple marker strategies to force a syntax error
	// Some markers work better in different contexts
	markers := []string{
		"@",  // Invalid character - should trigger lexer/parser error
		";",  // Valid token but often invalid in mid-query
		";;", // Double semicolon - definitely invalid
	}

	var predictListener *predictingErrorListener
	var p *parser.CqlParser

	for _, marker := range markers {
		queryWithMarker := truncated + marker

		// Create lexer
		inputStream := antlr.NewInputStream(queryWithMarker)
		lexer := parser.NewCqlLexer(inputStream)

		// Silence lexer errors
		lexer.RemoveErrorListeners()

		// Create token stream
		tokens := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)

		// Create parser
		p = parser.NewCqlParser(tokens)

		// Use our predicting error listener to capture expected tokens
		predictListener = &predictingErrorListener{
			DefaultErrorListener: antlr.NewDefaultErrorListener(),
		}
		p.RemoveErrorListeners()
		p.AddErrorListener(predictListener)

		// Enable error recovery to get as far as possible
		p.SetErrorHandler(antlr.NewDefaultErrorStrategy())

		// Parse - the marker should cause an error
		_ = p.Root()

		// If we captured expected tokens, we're done
		if predictListener.captured && predictListener.expectedTokens != nil {
			break
		}

		// Also try capturing from parser state after parsing
		predictListener.captureFromParser(p)
		if predictListener.captured && predictListener.expectedTokens != nil {
			break
		}
	}

	// Build result
	result := &PredictedTokens{
		ExpectedTokenTypes: make([]int, 0),
		TokenNames:         make(map[int]string),
	}

	// Get token names for mapping
	var symbolicNames, literalNames []string
	var expectedSet *antlr.IntervalSet

	if predictListener != nil && predictListener.captured && predictListener.expectedTokens != nil {
		expectedSet = predictListener.expectedTokens
		symbolicNames = predictListener.symbolicNames
		literalNames = predictListener.literalNames
	}

	// Convert IntervalSet to slice of token types
	if expectedSet != nil {
		for _, interval := range expectedSet.GetIntervals() {
			for tokenType := interval.Start; tokenType <= interval.Stop; tokenType++ {
				result.ExpectedTokenTypes = append(result.ExpectedTokenTypes, tokenType)

				// Get token name
				name := ""
				if tokenType >= 0 && tokenType < len(symbolicNames) && symbolicNames[tokenType] != "" {
					name = symbolicNames[tokenType]
				} else if tokenType >= 0 && tokenType < len(literalNames) && literalNames[tokenType] != "" {
					name = literalNames[tokenType]
				}
				result.TokenNames[tokenType] = name
			}
		}
	}

	return result
}

// IsKeywordToken returns true if the token type is a CQL keyword (K_* tokens).
func IsKeywordToken(tokenType int, tokenName string) bool {
	return strings.HasPrefix(tokenName, "K_")
}

// IsTokenValidAtPosition checks if inserting a specific token at the given position
// would result in a valid (or potentially valid) parse.
// This is useful for filtering completion suggestions - it validates by actually parsing.
func IsTokenValidAtPosition(query string, position int, token string) bool {
	// Truncate query at cursor position
	if position > len(query) {
		position = len(query)
	}
	truncated := query[:position]

	// Insert the token and try to parse
	// Add a placeholder after to make it more likely to detect issues
	testQuery := truncated + token + " __placeholder__"

	// Create lexer
	inputStream := antlr.NewInputStream(testQuery)
	lexer := parser.NewCqlLexer(inputStream)
	lexer.RemoveErrorListeners()

	// Create token stream
	tokens := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)

	// Create parser
	p := parser.NewCqlParser(tokens)

	// Track if we had a syntax error at or before the token position
	errorListener := &validationErrorListener{
		tokenEndPosition: position + len(token),
	}
	p.RemoveErrorListeners()
	p.AddErrorListener(errorListener)

	// Use bail error strategy to stop at first error
	p.SetErrorHandler(antlr.NewDefaultErrorStrategy())

	// Parse
	func() {
		defer func() {
			_ = recover() // Handle any panics
		}()
		_ = p.Root()
	}()

	// If no error occurred before the token end, the token is valid
	return !errorListener.hadEarlyError
}

// validationErrorListener tracks syntax errors for validation.
type validationErrorListener struct {
	*antlr.DefaultErrorListener
	tokenEndPosition int
	hadEarlyError    bool
}

func (v *validationErrorListener) SyntaxError(
	recognizer antlr.Recognizer,
	offendingSymbol any,
	line, column int,
	msg string,
	e antlr.RecognitionException,
) {
	// Check if the error is at or before our token position
	// Column is 0-indexed, and we need to account for line 1
	if line == 1 && column <= v.tokenEndPosition {
		v.hadEarlyError = true
	}
}

// TokenTypeToKeyword converts a K_* token name to a CQL keyword.
// Example: "K_SELECT" -> "SELECT", "K_WHERE" -> "WHERE"
func TokenTypeToKeyword(tokenName string) string {
	if strings.HasPrefix(tokenName, "K_") {
		return strings.ToUpper(tokenName[2:])
	}
	return tokenName
}

// FilterCompletionsWithANTLR filters a list of completion items using ANTLR validation.
// It returns only completions that would be grammatically valid at the given position.
// This is useful for removing false positives from heuristic-based completions.
func FilterCompletionsWithANTLR(query string, position int, items []CompletionItem) []CompletionItem {
	result := make([]CompletionItem, 0, len(items))
	for _, item := range items {
		// Use the label for validation
		token := item.Label
		if IsTokenValidAtPosition(query, position, token) {
			result = append(result, item)
		}
	}
	return result
}

// FilterKeywordsWithANTLR filters a list of keyword strings using ANTLR validation.
// Returns only keywords that would be grammatically valid at the given position.
func FilterKeywordsWithANTLR(query string, position int, keywords []string) []string {
	result := make([]string, 0, len(keywords))
	for _, kw := range keywords {
		if IsTokenValidAtPosition(query, position, kw) {
			result = append(result, kw)
		}
	}
	return result
}

// FilterKeywordTokens returns only the keyword tokens from the predicted set.
func FilterKeywordTokens(predicted *PredictedTokens) []string {
	keywords := make([]string, 0)
	seen := make(map[string]bool)

	for _, tokenType := range predicted.ExpectedTokenTypes {
		name := predicted.TokenNames[tokenType]
		if IsKeywordToken(tokenType, name) {
			keyword := TokenTypeToKeyword(name)
			if !seen[keyword] {
				seen[keyword] = true
				keywords = append(keywords, keyword)
			}
		}
	}

	return keywords
}
