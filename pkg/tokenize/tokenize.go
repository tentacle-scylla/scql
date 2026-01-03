// Package tokenize provides CQL tokenization for syntax highlighting.
package tokenize

import (
	"strings"

	"github.com/antlr4-go/antlr/v4"

	parser "github.com/pierre-borckmans/scql/gen/parser"
	"github.com/pierre-borckmans/scql/pkg/hover"
)

// TokenType identifies the semantic type of a token for syntax highlighting.
type TokenType string

const (
	TokenKeyword       TokenType = "keyword"
	TokenFunction      TokenType = "function"
	TokenType_         TokenType = "type"
	TokenString        TokenType = "string"
	TokenNumber        TokenType = "number"
	TokenComment       TokenType = "comment"
	TokenIdentifier    TokenType = "identifier"
	TokenOperator      TokenType = "operator"
	TokenPunctuation   TokenType = "punctuation"
	TokenPartitionKey  TokenType = "partition_key"
	TokenClusteringKey TokenType = "clustering_key"
	TokenColumn        TokenType = "column"
	TokenPlaceholder   TokenType = "placeholder"
)

// Token represents a single token for syntax highlighting.
type Token struct {
	Start int       `json:"start"`
	End   int       `json:"end"`
	Text  string    `json:"text"`
	Type  TokenType `json:"type"`
}

// Context provides semantic information for enhanced tokenization.
type Context struct {
	PartitionKeys  []string
	ClusteringKeys []string
	Columns        []string
}

// Tokenize returns all tokens from a CQL string with semantic classification.
func Tokenize(input string, ctx *Context) []Token {
	if input == "" {
		return nil
	}

	// Create lexer
	inputStream := antlr.NewInputStream(input)
	lexer := parser.NewCqlLexer(inputStream)

	// Disable error output
	lexer.RemoveErrorListeners()

	// Get all tokens
	tokens := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	tokens.Fill()

	allTokens := tokens.GetAllTokens()
	result := make([]Token, 0, len(allTokens))

	// Build lookup maps for semantic context
	var partitionKeyMap, clusteringKeyMap, columnMap map[string]bool
	if ctx != nil {
		partitionKeyMap = makeSet(ctx.PartitionKeys)
		clusteringKeyMap = makeSet(ctx.ClusteringKeys)
		columnMap = makeSet(ctx.Columns)
	}

	for _, tok := range allTokens {
		// Skip EOF token
		if tok.GetTokenType() == antlr.TokenEOF {
			continue
		}

		// Skip hidden channel tokens (whitespace) - GetAllTokens includes them
		if tok.GetChannel() == antlr.TokenHiddenChannel {
			continue
		}

		text := tok.GetText()
		start := tok.GetStart()
		end := tok.GetStop() + 1 // ANTLR stop is inclusive, we want exclusive

		// Classify the token
		tokenType := classifyANTLRToken(tok, input, partitionKeyMap, clusteringKeyMap, columnMap)

		result = append(result, Token{
			Start: start,
			End:   end,
			Text:  text,
			Type:  tokenType,
		})
	}

	return result
}

// classifyANTLRToken determines the semantic type of an ANTLR token.
func classifyANTLRToken(tok antlr.Token, input string, partitionKeys, clusteringKeys, columns map[string]bool) TokenType {
	tokenType := tok.GetTokenType()
	text := tok.GetText()
	lowerText := strings.ToLower(text)

	// Handle hidden channel tokens (comments, whitespace)
	if tok.GetChannel() == antlr.TokenHiddenChannel {
		// Check if it's a comment
		if strings.HasPrefix(text, "--") || strings.HasPrefix(text, "//") || strings.HasPrefix(text, "/*") {
			return TokenComment
		}
		// Skip whitespace (shouldn't happen as we use DefaultChannel, but be safe)
		return TokenIdentifier
	}

	// Check punctuation first
	switch tokenType {
	case parser.CqlLexerLR_BRACKET, parser.CqlLexerRR_BRACKET,
		parser.CqlLexerLC_BRACKET, parser.CqlLexerRC_BRACKET,
		parser.CqlLexerLS_BRACKET, parser.CqlLexerRS_BRACKET,
		parser.CqlLexerCOMMA, parser.CqlLexerSEMI, parser.CqlLexerCOLON, parser.CqlLexerDOT:
		return TokenPunctuation
	}

	// Check operators
	switch tokenType {
	case parser.CqlLexerSTAR, parser.CqlLexerDIVIDE, parser.CqlLexerMODULE,
		parser.CqlLexerPLUS, parser.CqlLexerMINUS, parser.CqlLexerMINUSMINUS,
		parser.CqlLexerOPERATOR_EQ, parser.CqlLexerOPERATOR_LT, parser.CqlLexerOPERATOR_GT,
		parser.CqlLexerOPERATOR_LTE, parser.CqlLexerOPERATOR_GTE, parser.CqlLexerOPERATOR_NEQ:
		return TokenOperator
	}

	// Check string literals
	if tokenType == parser.CqlLexerSTRING_LITERAL || tokenType == parser.CqlLexerCODE_BLOCK {
		return TokenString
	}

	// Check numeric literals
	switch tokenType {
	case parser.CqlLexerDECIMAL_LITERAL, parser.CqlLexerFLOAT_LITERAL,
		parser.CqlLexerHEXADECIMAL_LITERAL, parser.CqlLexerREAL_LITERAL:
		return TokenNumber
	}

	// Check placeholder
	if tokenType == parser.CqlLexerQMARK {
		return TokenPlaceholder
	}

	// Check UUID (treat as identifier/literal)
	if tokenType == parser.CqlLexerUUID {
		return TokenIdentifier
	}

	// Check duration literal
	if tokenType == parser.CqlLexerDURATION_LITERAL {
		return TokenNumber
	}

	// Check comments (should be on hidden channel, but check anyway)
	switch tokenType {
	case parser.CqlLexerCOMMENT_INPUT, parser.CqlLexerLINE_COMMENT, parser.CqlLexerSPEC_MYSQL_COMMENT:
		return TokenComment
	}

	// For identifiers and keywords, we need more sophisticated classification
	if tokenType == parser.CqlLexerOBJECT_NAME {
		// Check semantic context first
		if partitionKeys != nil && partitionKeys[lowerText] {
			return TokenPartitionKey
		}
		if clusteringKeys != nil && clusteringKeys[lowerText] {
			return TokenClusteringKey
		}
		if columns != nil && columns[lowerText] {
			return TokenColumn
		}
		return TokenIdentifier
	}

	// Check if it's a function (keyword followed by parenthesis)
	afterToken := strings.TrimSpace(input[tok.GetStop()+1:])
	if len(afterToken) > 0 && afterToken[0] == '(' {
		if hover.IsFunction(lowerText) {
			return TokenFunction
		}
		// Even if not a known function, treat as function call
		return TokenFunction
	}

	// Check if it's a type keyword
	if hover.IsType(lowerText) {
		return TokenType_
	}

	// Check if it's a reserved keyword.
	// Use hover.IsKeyword() which checks against the actual list of reserved keywords,
	// not just the ANTLR token type. Many K_* tokens (like K_USERS) are non-reserved
	// and commonly used as identifiers.
	if hover.IsKeyword(lowerText) {
		// If we have semantic context and this keyword matches a known column/key, treat it as such
		// (allows using reserved keywords as column names if quoted, which Cassandra supports)
		if partitionKeys != nil && partitionKeys[lowerText] {
			return TokenPartitionKey
		}
		if clusteringKeys != nil && clusteringKeys[lowerText] {
			return TokenClusteringKey
		}
		if columns != nil && columns[lowerText] {
			return TokenColumn
		}
		// It's a reserved keyword not matching any semantic context
		return TokenKeyword
	}

	// For ANTLR keywords that are NOT reserved (like USERS, STATUS, etc.),
	// check semantic context to classify them as columns/keys if applicable
	if isKeywordToken(tokenType) {
		if partitionKeys != nil && partitionKeys[lowerText] {
			return TokenPartitionKey
		}
		if clusteringKeys != nil && clusteringKeys[lowerText] {
			return TokenClusteringKey
		}
		if columns != nil && columns[lowerText] {
			return TokenColumn
		}
		// Non-reserved keyword used as identifier (e.g., table name "users")
		return TokenIdentifier
	}

	// Default to identifier
	return TokenIdentifier
}

// isKeywordToken returns true if the token type is a CQL keyword.
func isKeywordToken(tokenType int) bool {
	// K_ADD through K_VECTOR_SEARCH_INDEXING are all keywords
	// Based on the lexer, keywords start at K_ADD (24) and go through the various K_* tokens
	return tokenType >= parser.CqlLexerK_ADD && tokenType <= parser.CqlLexerK_VECTOR_SEARCH_INDEXING
}

// makeSet creates a case-insensitive lookup set from a slice of strings.
func makeSet(items []string) map[string]bool {
	if len(items) == 0 {
		return nil
	}
	m := make(map[string]bool, len(items))
	for _, item := range items {
		m[strings.ToLower(item)] = true
	}
	return m
}
