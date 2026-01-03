package parse

import (
	"os"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// ErrorFixture represents a single test case from the fixtures file
type ErrorFixture struct {
	Name                     string   `yaml:"name"`
	Query                    string   `yaml:"query"`
	ExpectError              bool     `yaml:"expectError"`
	ExpectFriendlyContains   []string `yaml:"expectFriendlyContains,omitempty"`
	ExpectSuggestionContains string   `yaml:"expectSuggestionContains,omitempty"`
	Comment                  string   `yaml:"comment,omitempty"`
}

// ErrorFixturesFile represents the structure of the fixtures YAML file
type ErrorFixturesFile struct {
	Description string         `yaml:"description"`
	Fixtures    []ErrorFixture `yaml:"fixtures"`
}

func loadFixtures(t *testing.T) []ErrorFixture {
	t.Helper()

	data, err := os.ReadFile("testdata/error_fixtures.yaml")
	if err != nil {
		t.Fatalf("Failed to load fixtures: %v", err)
	}

	var file ErrorFixturesFile
	if err := yaml.Unmarshal(data, &file); err != nil {
		t.Fatalf("Failed to parse fixtures: %v", err)
	}

	return file.Fixtures
}

func TestErrorFixtures(t *testing.T) {
	fixtures := loadFixtures(t)

	for _, f := range fixtures {
		t.Run(f.Name, func(t *testing.T) {
			result := Parse(f.Query)

			// Check if error is expected
			if f.ExpectError {
				if !result.HasErrors() {
					t.Errorf("Expected error for query: %s", f.Query)
					return
				}

				err := result.Errors.First()

				// Check friendly message contains all expected substrings
				if len(f.ExpectFriendlyContains) > 0 {
					friendly := err.FriendlyMessage
					if friendly == "" {
						friendly = err.Message
					}
					for _, expected := range f.ExpectFriendlyContains {
						if !strings.Contains(strings.ToUpper(friendly), strings.ToUpper(expected)) {
							t.Errorf("FriendlyMessage %q should contain %q", friendly, expected)
						}
					}
				}

				// Check suggestion contains expected text
				if f.ExpectSuggestionContains != "" {
					if err.Suggestion == "" {
						t.Errorf("Expected suggestion containing %q but got none", f.ExpectSuggestionContains)
					} else if !strings.Contains(strings.ToUpper(err.Suggestion), strings.ToUpper(f.ExpectSuggestionContains)) {
						t.Errorf("Suggestion %q should contain %q", err.Suggestion, f.ExpectSuggestionContains)
					}
				}

				// Log the error details for debugging
				t.Logf("Query: %s", f.Query)
				t.Logf("Raw message: %s", err.Message)
				t.Logf("Friendly: %s", err.FriendlyMessage)
				t.Logf("Suggestion: %s", err.Suggestion)
			} else {
				if result.HasErrors() {
					t.Errorf("Expected no error for query %q, got: %v", f.Query, result.Errors)
				}
			}
		})
	}
}

func TestTransformError(t *testing.T) {
	tests := []struct {
		name           string
		rawMessage     string
		query          string
		wantFriendly   string
		wantSuggestion string
	}{
		{
			name:           "no viable alternative with FORM typo",
			rawMessage:     "no viable alternative at input 'FORM'",
			query:          "SELECT * FORM users",
			wantFriendly:   "Unknown keyword 'FORM'",
			wantSuggestion: "Did you mean 'FROM'?",
		},
		{
			name:         "generic no viable alternative",
			rawMessage:   "no viable alternative at input 'xyz'",
			query:        "SELECT xyz FROM users",
			wantFriendly: "Unexpected syntax near 'xyz'",
		},
		{
			name:         "extraneous input",
			rawMessage:   "extraneous input 'FROM' expecting",
			query:        "SELECT * FROM FROM users",
			wantFriendly: "Unexpected 'FROM'",
		},
		{
			name:         "mismatched input",
			rawMessage:   "mismatched input ';' expecting something",
			query:        "SELECT * FROM;",
			wantFriendly: "Unexpected ';'",
		},
		{
			name:         "missing keyword",
			rawMessage:   "missing 'FROM' at 'users'",
			query:        "SELECT * users",
			wantFriendly: "Missing 'FROM' before 'users'",
		},
		{
			name:         "token recognition error",
			rawMessage:   "token recognition error at: '@'",
			query:        "SELECT @ FROM users",
			wantFriendly: "Invalid character '@'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TransformError(tt.rawMessage, tt.query)

			if tt.wantFriendly != "" && result.FriendlyMessage != tt.wantFriendly {
				t.Errorf("FriendlyMessage = %q, want %q", result.FriendlyMessage, tt.wantFriendly)
			}

			if tt.wantSuggestion != "" && result.Suggestion != tt.wantSuggestion {
				t.Errorf("Suggestion = %q, want %q", result.Suggestion, tt.wantSuggestion)
			}
		})
	}
}

func TestLevenshteinDistance(t *testing.T) {
	tests := []struct {
		s1, s2 string
		want   int
	}{
		{"", "", 0},
		{"a", "", 1},
		{"", "a", 1},
		{"abc", "abc", 0},
		{"abc", "ab", 1},
		{"abc", "abcd", 1},
		{"abc", "adc", 1},
		{"FORM", "FROM", 2},
		{"SELEC", "SELECT", 1},
		{"WHER", "WHERE", 1},
		{"kitten", "sitting", 3},
	}

	for _, tt := range tests {
		t.Run(tt.s1+"_"+tt.s2, func(t *testing.T) {
			got := levenshteinDistance(tt.s1, tt.s2)
			if got != tt.want {
				t.Errorf("levenshteinDistance(%q, %q) = %d, want %d", tt.s1, tt.s2, got, tt.want)
			}
		})
	}
}

func TestSuggestKeyword(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"FORM", "FROM"},
		{"FRMO", "FROM"},
		{"SELEC", "SELECT"},
		{"SLECT", "SELECT"},
		{"WHER", "WHERE"},
		{"WHRE", "WHERE"},
		{"INSRT", "INSERT"},
		{"UDPATE", "UPDATE"},
		{"DELTE", "DELETE"},
		{"CRETAE", "CREATE"},
		{"TABL", "TABLE"},
		{"PRIMRY", "PRIMARY"},
		{"VALUS", "VALUES"},
		{"LIMTI", "LIMIT"},
		{"ORDR", "ORDER"},
		{"xyz", ""},    // Too far from any keyword
		{"BANANA", ""}, // Not close to any keyword
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := SuggestKeyword(tt.input)
			if got != tt.want {
				t.Errorf("SuggestKeyword(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestErrorPatternCoverage(t *testing.T) {
	// Test that all patterns in the registry are valid
	for _, pattern := range errorPatterns {
		if pattern.Name == "" {
			t.Error("Pattern has empty name")
		}
		if pattern.MessagePattern == nil {
			t.Errorf("Pattern %q has nil MessagePattern", pattern.Name)
		}
		if pattern.FriendlyTemplate == "" {
			t.Errorf("Pattern %q has empty FriendlyTemplate", pattern.Name)
		}
	}
}
