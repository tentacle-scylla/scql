package lint

import (
	"testing"

	"github.com/pierre-borckmans/scql/pkg/types"
)

func TestCheck(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantErrors bool
	}{
		{
			name:       "valid query",
			input:      "SELECT * FROM users;",
			wantErrors: false,
		},
		{
			name:       "invalid query",
			input:      "SELEC * FROM users;",
			wantErrors: true,
		},
		{
			name:       "typo in keyword",
			input:      "SELECT * FORM users;",
			wantErrors: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := Check(tt.input)
			if errors.HasErrors() != tt.wantErrors {
				t.Errorf("HasErrors() = %v, want %v", errors.HasErrors(), tt.wantErrors)
			}
		})
	}
}

func TestCheckMultiple(t *testing.T) {
	input := `
		SELECT * FROM users;
		SELEC * FROM invalid;
		INSERT INTO users (id) VALUES (1);
	`

	errors := CheckMultiple(input)

	if !errors.HasErrors() {
		t.Error("expected errors but got none")
	}

	if len(errors) != 1 {
		t.Errorf("got %d errors, want 1", len(errors))
	}
}

func TestIsValid(t *testing.T) {
	if !IsValid("SELECT * FROM users;") {
		t.Error("valid query should return true")
	}

	if IsValid("INVALID QUERY;") {
		t.Error("invalid query should return false")
	}
}

func TestAnalyze(t *testing.T) {
	result := Analyze("SELECT * FROM users WHERE id = 1;")

	if !result.IsValid {
		t.Error("Expected valid result")
	}

	if result.Type != types.StatementSelect {
		t.Errorf("Type = %v, want SELECT", result.Type)
	}

	if result.Errors.HasErrors() {
		t.Error("Expected no errors")
	}
}

func TestAnalyzeInvalid(t *testing.T) {
	result := Analyze("SELECT * FORM users;")

	if result.IsValid {
		t.Error("Expected invalid result")
	}

	if !result.Errors.HasErrors() {
		t.Error("Expected errors")
	}

	// Check that error has position
	err := result.Errors.First()
	if err.Line != 1 {
		t.Errorf("Line = %d, want 1", err.Line)
	}
}

func TestAnalyzeMultiple(t *testing.T) {
	input := `
		SELECT * FROM users;
		INSERT INTO users (id) VALUES (1);
		UPDATE users SET name = 'test' WHERE id = 1;
	`

	results := AnalyzeMultiple(input)

	if len(results) != 3 {
		t.Errorf("got %d results, want 3", len(results))
	}

	expectedTypes := []types.StatementType{
		types.StatementSelect,
		types.StatementInsert,
		types.StatementUpdate,
	}

	for i, r := range results {
		if !r.IsValid {
			t.Errorf("result[%d] is not valid", i)
		}
		if r.Type != expectedTypes[i] {
			t.Errorf("result[%d].Type = %v, want %v", i, r.Type, expectedTypes[i])
		}
	}
}
