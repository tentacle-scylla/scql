package analyze

import (
	"os"
	"strings"
	"testing"

	"github.com/tentacle-scylla/scql/pkg/schema"
	"gopkg.in/yaml.v3"
)

// FixtureColumn represents a column in the fixture schema
type FixtureColumn struct {
	Name string `yaml:"name"`
	Type string `yaml:"type"`
}

// FixtureTable represents a table in the fixture schema
type FixtureTable struct {
	Name          string          `yaml:"name"`
	PartitionKey  []string        `yaml:"partitionKey"`
	ClusteringKey []string        `yaml:"clusteringKey"`
	Columns       []FixtureColumn `yaml:"columns"`
}

// FixtureKeyspace represents a keyspace in the fixture schema
type FixtureKeyspace struct {
	Name   string         `yaml:"name"`
	Tables []FixtureTable `yaml:"tables"`
}

// FixtureSchema represents the schema in a fixture
type FixtureSchema struct {
	Keyspaces []FixtureKeyspace `yaml:"keyspaces"`
}

// FixtureOptions represents analyze options in a fixture
type FixtureOptions struct {
	DefaultKeyspace     string `yaml:"defaultKeyspace"`
	WarnOnSelectStar    bool   `yaml:"warnOnSelectStar"`
	WarnOnNoLimit       bool   `yaml:"warnOnNoLimit"`
	LargeLimitThreshold int    `yaml:"largeLimitThreshold"`
}

// AnalyzeFixture represents a single test case
type AnalyzeFixture struct {
	Name    string `yaml:"name"`
	Query   string `yaml:"query"`
	Comment string `yaml:"comment,omitempty"`

	// Schema reference (by name from schemas section)
	SchemaRef string          `yaml:"schemaRef,omitempty"`
	Options   *FixtureOptions `yaml:"options,omitempty"`

	// Reference extraction expectations
	ExpectKeyspace       string   `yaml:"expectKeyspace,omitempty"`
	ExpectTable          string   `yaml:"expectTable,omitempty"`
	ExpectColumns        []string `yaml:"expectColumns,omitempty"`
	ExpectSelectColumns  []string `yaml:"expectSelectColumns,omitempty"`
	ExpectWhereColumns   []string `yaml:"expectWhereColumns,omitempty"`
	ExpectUpdateColumns  []string `yaml:"expectUpdateColumns,omitempty"`
	ExpectInsertColumns  []string `yaml:"expectInsertColumns,omitempty"`
	ExpectOrderByColumns []string `yaml:"expectOrderByColumns,omitempty"`
	ExpectFunctions      []string `yaml:"expectFunctions,omitempty"`
	ExpectLimit          *int     `yaml:"expectLimit,omitempty"`
	ExpectHasAllowFilter bool     `yaml:"expectHasAllowFiltering,omitempty"`

	// Validation expectations
	ExpectValid            *bool  `yaml:"expectValid,omitempty"`
	ExpectSyntaxError      bool   `yaml:"expectSyntaxError,omitempty"`
	ExpectSchemaErrorCount *int   `yaml:"expectSchemaErrorCount,omitempty"`
	ExpectSchemaErrorType  string `yaml:"expectSchemaErrorType,omitempty"`
	ExpectSchemaErrorCont  string `yaml:"expectSchemaErrorContains,omitempty"`
	ExpectSuggestionCont   string `yaml:"expectSuggestionContains,omitempty"`

	// Warning expectations
	ExpectWarningCount    *int   `yaml:"expectWarningCount,omitempty"`
	ExpectWarningType     string `yaml:"expectWarningType,omitempty"`
	ExpectWarningContains string `yaml:"expectWarningContains,omitempty"`
}

// FixtureFile represents the entire fixture file structure
type FixtureFile struct {
	Schemas map[string]FixtureSchema `yaml:"schemas"`
	Tests   []AnalyzeFixture         `yaml:"tests"`
}

func loadFixtures(t *testing.T) (*FixtureFile, error) {
	data, err := os.ReadFile("testdata/analyze_fixtures.yaml")
	if err != nil {
		return nil, err
	}

	var ff FixtureFile
	if err := yaml.Unmarshal(data, &ff); err != nil {
		return nil, err
	}

	return &ff, nil
}

func buildSchema(fs *FixtureSchema) *schema.Schema {
	if fs == nil {
		return nil
	}

	s := schema.NewSchema()
	for _, fks := range fs.Keyspaces {
		ks := s.AddKeyspace(fks.Name)
		for _, ftbl := range fks.Tables {
			tbl := ks.AddTable(ftbl.Name)

			// Add columns
			for _, fcol := range ftbl.Columns {
				tbl.AddColumn(fcol.Name, fcol.Type)
			}

			// Set partition key
			if len(ftbl.PartitionKey) > 0 {
				tbl.SetPartitionKey(ftbl.PartitionKey...)
			}

			// Set clustering key
			if len(ftbl.ClusteringKey) > 0 {
				tbl.SetClusteringKey(ftbl.ClusteringKey...)
			}
		}
	}
	return s
}

func buildOptions(fo *FixtureOptions, s *schema.Schema) *AnalyzeOptions {
	opts := DefaultOptions()
	opts.Schema = s

	if fo != nil {
		opts.DefaultKeyspace = fo.DefaultKeyspace
		opts.WarnOnSelectStar = fo.WarnOnSelectStar
		opts.WarnOnNoLimit = fo.WarnOnNoLimit
		opts.LargeLimitThreshold = fo.LargeLimitThreshold
	}

	return opts
}

func TestAnalyzeFixtures(t *testing.T) {
	ff, err := loadFixtures(t)
	if err != nil {
		t.Fatalf("Failed to load fixtures: %v", err)
	}

	for _, f := range ff.Tests {
		t.Run(f.Name, func(t *testing.T) {
			// Look up schema by reference
			var s *schema.Schema
			if f.SchemaRef != "" {
				fs, ok := ff.Schemas[f.SchemaRef]
				if !ok {
					t.Fatalf("Schema reference %q not found", f.SchemaRef)
				}
				s = buildSchema(&fs)
			}

			opts := buildOptions(f.Options, s)
			result := Analyze(f.Query, opts)

			// Log for debugging
			t.Logf("Query: %s", f.Query)
			if result.References != nil {
				t.Logf("Refs: keyspace=%q table=%q", result.References.Keyspace, result.References.Table)
			}
			if len(result.SyntaxErrors) > 0 {
				t.Logf("Syntax errors: %d", len(result.SyntaxErrors))
			}
			if len(result.SchemaErrors) > 0 {
				t.Logf("Schema errors: %d", len(result.SchemaErrors))
				for _, e := range result.SchemaErrors {
					t.Logf("  - %s: %s (suggestion: %s)", e.Type, e.Message, e.Suggestion)
				}
			}
			if len(result.Warnings) > 0 {
				t.Logf("Warnings: %d", len(result.Warnings))
				for _, w := range result.Warnings {
					t.Logf("  - %s: %s", w.Type, w.Message)
				}
			}

			// Check syntax error expectation
			if f.ExpectSyntaxError {
				if len(result.SyntaxErrors) == 0 {
					t.Error("Expected syntax error but got none")
				}
			}

			// Check validity
			if f.ExpectValid != nil {
				if result.IsValid != *f.ExpectValid {
					t.Errorf("IsValid = %v, want %v", result.IsValid, *f.ExpectValid)
				}
			}

			// Check reference extraction
			refs := result.References

			if f.ExpectKeyspace != "" && refs.Keyspace != f.ExpectKeyspace {
				t.Errorf("Keyspace = %q, want %q", refs.Keyspace, f.ExpectKeyspace)
			}

			if f.ExpectTable != "" && refs.Table != f.ExpectTable {
				t.Errorf("Table = %q, want %q", refs.Table, f.ExpectTable)
			}

			if len(f.ExpectColumns) > 0 {
				checkStringSlice(t, "Columns", refs.Columns, f.ExpectColumns)
			}

			if len(f.ExpectSelectColumns) > 0 {
				checkStringSlice(t, "SelectColumns", refs.SelectColumns, f.ExpectSelectColumns)
			}

			if len(f.ExpectWhereColumns) > 0 {
				checkStringSlice(t, "WhereColumns", refs.WhereColumns, f.ExpectWhereColumns)
			}

			if len(f.ExpectUpdateColumns) > 0 {
				checkStringSlice(t, "UpdateColumns", refs.UpdateColumns, f.ExpectUpdateColumns)
			}

			if len(f.ExpectInsertColumns) > 0 {
				checkStringSlice(t, "InsertColumns", refs.InsertColumns, f.ExpectInsertColumns)
			}

			if len(f.ExpectOrderByColumns) > 0 {
				checkStringSlice(t, "OrderByColumns", refs.OrderByColumns, f.ExpectOrderByColumns)
			}

			if len(f.ExpectFunctions) > 0 {
				checkStringSlice(t, "Functions", refs.Functions, f.ExpectFunctions)
			}

			if f.ExpectLimit != nil && refs.Limit != *f.ExpectLimit {
				t.Errorf("Limit = %d, want %d", refs.Limit, *f.ExpectLimit)
			}

			if f.ExpectHasAllowFilter && !refs.HasAllowFiltering {
				t.Error("Expected HasAllowFiltering = true")
			}

			// Check schema errors
			if f.ExpectSchemaErrorCount != nil {
				if len(result.SchemaErrors) != *f.ExpectSchemaErrorCount {
					t.Errorf("SchemaError count = %d, want %d", len(result.SchemaErrors), *f.ExpectSchemaErrorCount)
				}
			}

			if f.ExpectSchemaErrorType != "" {
				found := false
				for _, e := range result.SchemaErrors {
					if string(e.Type) == f.ExpectSchemaErrorType {
						found = true

						if f.ExpectSchemaErrorCont != "" && !strings.Contains(e.Message, f.ExpectSchemaErrorCont) {
							t.Errorf("Schema error message %q does not contain %q", e.Message, f.ExpectSchemaErrorCont)
						}

						if f.ExpectSuggestionCont != "" && !strings.Contains(e.Suggestion, f.ExpectSuggestionCont) {
							t.Errorf("Suggestion %q does not contain %q", e.Suggestion, f.ExpectSuggestionCont)
						}
						break
					}
				}
				if !found {
					t.Errorf("Expected schema error of type %q but not found", f.ExpectSchemaErrorType)
				}
			}

			// Check warnings
			if f.ExpectWarningCount != nil {
				if len(result.Warnings) != *f.ExpectWarningCount {
					t.Errorf("Warning count = %d, want %d", len(result.Warnings), *f.ExpectWarningCount)
				}
			}

			if f.ExpectWarningType != "" {
				found := false
				for _, w := range result.Warnings {
					if string(w.Type) == f.ExpectWarningType {
						found = true

						if f.ExpectWarningContains != "" && !strings.Contains(w.Message, f.ExpectWarningContains) {
							t.Errorf("Warning message %q does not contain %q", w.Message, f.ExpectWarningContains)
						}
						break
					}
				}
				if !found {
					t.Errorf("Expected warning of type %q but not found", f.ExpectWarningType)
				}
			}
		})
	}
}

func checkStringSlice(t *testing.T, name string, got, want []string) {
	t.Helper()

	// Check that all expected items are present (order doesn't matter)
	for _, w := range want {
		found := false
		for _, g := range got {
			if g == w {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("%s missing expected item %q, got %v", name, w, got)
		}
	}
}
