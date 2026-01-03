package parse

import (
	"strings"
	"testing"

	"github.com/pierre-borckmans/scql/pkg/types"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantType  types.StatementType
		wantValid bool
	}{
		{
			name:      "simple select",
			input:     "SELECT * FROM users;",
			wantType:  types.StatementSelect,
			wantValid: true,
		},
		{
			name:      "select with where",
			input:     "SELECT id, name FROM users WHERE id = 1;",
			wantType:  types.StatementSelect,
			wantValid: true,
		},
		{
			name:      "insert",
			input:     "INSERT INTO users (id, name) VALUES (1, 'test');",
			wantType:  types.StatementInsert,
			wantValid: true,
		},
		{
			name:      "update",
			input:     "UPDATE users SET name = 'new' WHERE id = 1;",
			wantType:  types.StatementUpdate,
			wantValid: true,
		},
		{
			name:      "delete",
			input:     "DELETE FROM users WHERE id = 1;",
			wantType:  types.StatementDelete,
			wantValid: true,
		},
		{
			name:      "create table",
			input:     "CREATE TABLE users (id int PRIMARY KEY, name text);",
			wantType:  types.StatementCreateTable,
			wantValid: true,
		},
		{
			name:      "create keyspace",
			input:     "CREATE KEYSPACE test WITH REPLICATION = {'class': 'SimpleStrategy', 'replication_factor': 1};",
			wantType:  types.StatementCreateKeyspace,
			wantValid: true,
		},
		{
			name:      "invalid - typo in FROM",
			input:     "SELECT * FORM users;",
			wantType:  types.StatementUnknown,
			wantValid: false,
		},
		{
			name:      "without semicolon - still valid",
			input:     "SELECT * FROM users",
			wantType:  types.StatementSelect,
			wantValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Parse(tt.input)

			if result.IsValid() != tt.wantValid {
				t.Errorf("IsValid() = %v, want %v", result.IsValid(), tt.wantValid)
				if result.HasErrors() {
					for _, err := range result.Errors {
						t.Logf("  Error: %s", err.Error())
					}
				}
			}

			if tt.wantValid && result.Type != tt.wantType {
				t.Errorf("Type = %v, want %v", result.Type, tt.wantType)
			}
		})
	}
}

func TestMultiple(t *testing.T) {
	input := `
		SELECT * FROM users;
		INSERT INTO users (id, name) VALUES (1, 'test');
		UPDATE users SET name = 'new' WHERE id = 1;
	`

	results := Multiple(input)

	if len(results) != 3 {
		t.Errorf("got %d results, want 3", len(results))
	}

	expectedTypes := []types.StatementType{types.StatementSelect, types.StatementInsert, types.StatementUpdate}
	for i, r := range results {
		if !r.IsValid() {
			t.Errorf("result[%d] is not valid", i)
			continue
		}
		if r.Type != expectedTypes[i] {
			t.Errorf("result[%d].Type = %v, want %v", i, r.Type, expectedTypes[i])
		}
	}
}

func TestMultipleWithStrings(t *testing.T) {
	input := `INSERT INTO users (id, name) VALUES (1, 'hello; world');`

	results := Multiple(input)

	if len(results) != 1 {
		t.Errorf("got %d results, want 1 (semicolon in string should not split)", len(results))
	}

	if len(results) > 0 && !results[0].IsValid() {
		t.Errorf("result is not valid: %v", results[0].Errors)
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

func TestErrorPositions(t *testing.T) {
	result := Parse("SELECT * FORM users;")

	if !result.HasErrors() {
		t.Fatal("expected errors")
	}

	err := result.Errors.First()
	if err.Line != 1 {
		t.Errorf("Line = %d, want 1", err.Line)
	}

	if err.Column < 9 {
		t.Errorf("Column = %d, expected >= 9 (pointing to FORM)", err.Column)
	}
}

func TestErrorSuggestions(t *testing.T) {
	result := Parse("SELECT * FORM users;")

	if !result.HasErrors() {
		t.Fatal("expected errors")
	}

	err := result.Errors.First()
	if err.Suggestion == "" {
		t.Log("No suggestion provided (this is acceptable)")
	} else {
		t.Logf("Suggestion: %s", err.Suggestion)
		if !strings.Contains(strings.ToLower(err.Suggestion), "from") {
			t.Errorf("Expected suggestion to mention FROM")
		}
	}
}

func TestScyllaDBFeatures(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "BYPASS CACHE",
			input: "SELECT * FROM users BYPASS CACHE;",
		},
		{
			name:  "USING TIMEOUT",
			input: "SELECT * FROM users USING TIMEOUT 5000;",
		},
		{
			name:  "PER PARTITION LIMIT",
			input: "SELECT * FROM users PER PARTITION LIMIT 10;",
		},
		{
			name:  "GROUP BY",
			input: "SELECT user_id, count(*) FROM events GROUP BY user_id;",
		},
		{
			name:  "PRUNE MATERIALIZED VIEW",
			input: "PRUNE MATERIALIZED VIEW my_view;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Parse(tt.input)
			if !result.IsValid() {
				t.Errorf("Expected valid, got errors: %v", result.Errors)
			}
		})
	}
}

func TestComplexQueries(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "create table with all options",
			input: `CREATE TABLE IF NOT EXISTS ks.users (
				id uuid,
				name text,
				email text,
				created_at timestamp,
				PRIMARY KEY (id)
			) WITH comment = 'User table'
			AND compaction = {'class': 'LeveledCompactionStrategy'};`,
		},
		{
			name: "batch statement",
			input: `BEGIN BATCH
				INSERT INTO users (id, name) VALUES (uuid(), 'test');
				UPDATE users SET name = 'updated' WHERE id = 123e4567-e89b-12d3-a456-426614174000;
			APPLY BATCH;`,
		},
		{
			name: "insert with TTL and timestamp",
			input: `INSERT INTO users (id, name) VALUES (1, 'test')
				USING TTL 3600 AND TIMESTAMP 1234567890;`,
		},
		{
			name:  "update with IF condition",
			input: "UPDATE users SET name = 'new' WHERE id = 1 IF name = 'old';",
		},
		{
			name:  "delete with IF EXISTS",
			input: "DELETE FROM users WHERE id = 1 IF EXISTS;",
		},
		{
			name:  "select with CAST",
			input: "SELECT CAST(id AS text) FROM users;",
		},
		{
			name:  "select with JSON",
			input: "SELECT JSON * FROM users;",
		},
		{
			name:  "prepared statement placeholder",
			input: "SELECT * FROM users WHERE id = ?;",
		},
		{
			name:  "named placeholder",
			input: "SELECT * FROM users WHERE id = :user_id;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Parse(tt.input)
			if !result.IsValid() {
				t.Errorf("Expected valid, got errors: %v", result.Errors)
			}
		})
	}
}
