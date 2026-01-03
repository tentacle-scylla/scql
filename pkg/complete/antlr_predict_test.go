package complete

import (
	"testing"
)

func TestGetExpectedTokensAtPosition(t *testing.T) {
	// Note: ANTLR's GetExpectedTokens() returns ALL grammatically valid tokens.
	// Since CQL allows most keywords as identifiers, we get many tokens.
	// This test verifies the API works and captures tokens, not that it's "smart".
	tests := []struct {
		name               string
		query              string
		position           int
		minExpectedTokens  int  // Minimum expected token count
		expectSomeKeywords bool // Whether we expect to get any keywords
	}{
		{
			name:               "empty query - may not capture from lexer error",
			query:              "",
			position:           0,
			minExpectedTokens:  0, // May be 0 due to lexer error handling
			expectSomeKeywords: false,
		},
		{
			name:               "after SELECT - expects many tokens including keywords",
			query:              "SELECT ",
			position:           7,
			minExpectedTokens:  10, // CQL allows many tokens here
			expectSomeKeywords: true,
		},
		{
			name:               "after FROM - expects identifier or keywords-as-identifiers",
			query:              "SELECT * FROM ",
			position:           14,
			minExpectedTokens:  10,
			expectSomeKeywords: true,
		},
		{
			name:               "after WHERE - expects column names or keywords-as-columns",
			query:              "SELECT * FROM users WHERE ",
			position:           26,
			minExpectedTokens:  10,
			expectSomeKeywords: true,
		},
		{
			name:               "complete query - may not capture tokens",
			query:              "SELECT * FROM users LIMIT 10 ",
			position:           29,
			minExpectedTokens:  0, // Complete queries may return 0
			expectSomeKeywords: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			predicted := GetExpectedTokensAtPosition(tt.query, tt.position)
			keywords := FilterKeywordTokens(predicted)

			t.Logf("Query: %q (pos %d)", tt.query, tt.position)
			t.Logf("Expected tokens: %d", len(predicted.ExpectedTokenTypes))
			t.Logf("Keywords: %v", keywords)

			// Check we got at least minimum expected tokens
			if len(predicted.ExpectedTokenTypes) < tt.minExpectedTokens {
				t.Errorf("Expected at least %d tokens, got %d", tt.minExpectedTokens, len(predicted.ExpectedTokenTypes))
			}

			// Check keyword expectation
			if tt.expectSomeKeywords && len(keywords) == 0 {
				t.Errorf("Expected some keywords but got none")
			}
		})
	}
}

func TestFilterKeywordTokens(t *testing.T) {
	// Test the token name to keyword conversion
	tests := []struct {
		tokenName string
		expected  string
	}{
		{"K_SELECT", "SELECT"},
		{"K_WHERE", "WHERE"},
		{"K_FROM", "FROM"},
		{"K_ORDER", "ORDER"},
		{"OBJECT_NAME", "OBJECT_NAME"}, // Non-keyword stays as-is
	}

	for _, tt := range tests {
		result := TokenTypeToKeyword(tt.tokenName)
		if result != tt.expected {
			t.Errorf("TokenTypeToKeyword(%q) = %q, want %q", tt.tokenName, result, tt.expected)
		}
	}
}

func TestIsTokenValidAtPosition(t *testing.T) {
	// Test the validation function
	tests := []struct {
		name     string
		query    string
		position int
		token    string
		expected bool
	}{
		{
			name:     "SELECT valid at empty query start",
			query:    "",
			position: 0,
			token:    "SELECT",
			expected: true,
		},
		{
			name:     "INSERT valid at empty query start",
			query:    "",
			position: 0,
			token:    "INSERT",
			expected: true,
		},
		{
			name:     "WHERE invalid at empty query start",
			query:    "",
			position: 0,
			token:    "WHERE",
			expected: false,
		},
		{
			name:     "FROM valid after SELECT columns",
			query:    "SELECT * ",
			position: 9,
			token:    "FROM",
			expected: true,
		},
		{
			name:     "WHERE valid after FROM table",
			query:    "SELECT * FROM users ",
			position: 20,
			token:    "WHERE",
			expected: true,
		},
		{
			name:     "WHERE invalid after LIMIT number",
			query:    "SELECT * FROM users LIMIT 10 ",
			position: 29,
			token:    "WHERE",
			expected: false,
		},
		{
			name:     "ORDER invalid after LIMIT number",
			query:    "SELECT * FROM users LIMIT 10 ",
			position: 29,
			token:    "ORDER",
			expected: false,
		},
		{
			name:     "ALLOW valid after LIMIT number",
			query:    "SELECT * FROM users LIMIT 10 ",
			position: 29,
			token:    "ALLOW",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsTokenValidAtPosition(tt.query, tt.position, tt.token)
			if result != tt.expected {
				t.Errorf("IsTokenValidAtPosition(%q, %d, %q) = %v, want %v",
					tt.query, tt.position, tt.token, result, tt.expected)
			}
		})
	}
}
