package complete

import (
	"os"
	"testing"

	"github.com/pierre-borckmans/scql/pkg/schema"
	"gopkg.in/yaml.v3"
)

// FixtureColumn represents a column in the fixture schema
type FixtureColumn struct {
	Name string `yaml:"name"`
	Type string `yaml:"type"`
}

// FixtureMV represents a materialized view in the fixture schema
type FixtureMV struct {
	Name          string          `yaml:"name"`
	BaseTable     string          `yaml:"baseTable"`
	PartitionKey  []string        `yaml:"partitionKey"`
	ClusteringKey []string        `yaml:"clusteringKey"`
	Columns       []FixtureColumn `yaml:"columns"`
}

// FixtureTable represents a table in the fixture schema
type FixtureTable struct {
	Name              string          `yaml:"name"`
	PartitionKey      []string        `yaml:"partitionKey"`
	ClusteringKey     []string        `yaml:"clusteringKey"`
	Columns           []FixtureColumn `yaml:"columns"`
	MaterializedViews []FixtureMV     `yaml:"materializedViews,omitempty"`
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

// CompleteFixture represents a single test case
type CompleteFixture struct {
	Name    string `yaml:"name"`
	Query   string `yaml:"query"`
	Position int   `yaml:"position"`
	Comment  string `yaml:"comment,omitempty"`

	// Schema context
	SchemaRef       string `yaml:"schemaRef,omitempty"`
	DefaultKeyspace string `yaml:"defaultKeyspace,omitempty"`
	TableContext    string `yaml:"tableContext,omitempty"`

	// Context detection expectations
	ExpectContext string `yaml:"expectContext,omitempty"`
	ExpectPrefix  string `yaml:"expectPrefix,omitempty"`

	// Completion expectations
	ExpectCompletionLabels []string `yaml:"expectCompletionLabels,omitempty"`
	ExpectCompletionKinds  []string `yaml:"expectCompletionKinds,omitempty"`
	ExpectMissingLabels    []string `yaml:"expectMissingLabels,omitempty"`
	ExpectFirstCompletions []string `yaml:"expectFirstCompletions,omitempty"`
}

// FixtureFile represents the entire fixture file structure
type FixtureFile struct {
	Schemas map[string]FixtureSchema `yaml:"schemas"`
	Tests   []CompleteFixture        `yaml:"tests"`
}

func loadFixtures(t *testing.T) (*FixtureFile, error) {
	data, err := os.ReadFile("testdata/complete_fixtures.yaml")
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

			for _, fcol := range ftbl.Columns {
				tbl.AddColumn(fcol.Name, fcol.Type)
			}

			if len(ftbl.PartitionKey) > 0 {
				tbl.SetPartitionKey(ftbl.PartitionKey...)
			}

			if len(ftbl.ClusteringKey) > 0 {
				tbl.SetClusteringKey(ftbl.ClusteringKey...)
			}

			// Add materialized views
			for _, fmv := range ftbl.MaterializedViews {
				mv := tbl.AddMaterializedView(fmv.Name)
				for _, fcol := range fmv.Columns {
					mv.AddColumn(fcol.Name, fcol.Type)
				}
				if len(fmv.PartitionKey) > 0 {
					mv.SetPartitionKey(fmv.PartitionKey...)
				}
				if len(fmv.ClusteringKey) > 0 {
					mv.SetClusteringKey(fmv.ClusteringKey...)
				}
			}
		}
	}
	return s
}

func TestCompleteFixtures(t *testing.T) {
	ff, err := loadFixtures(t)
	if err != nil {
		t.Fatalf("Failed to load fixtures: %v", err)
	}

	for _, f := range ff.Tests {
		t.Run(f.Name, func(t *testing.T) {
			// Build schema if referenced
			var s *schema.Schema
			if f.SchemaRef != "" {
				fs, ok := ff.Schemas[f.SchemaRef]
				if !ok {
					t.Fatalf("Schema reference %q not found", f.SchemaRef)
				}
				s = buildSchema(&fs)
			}

			// Test context detection
			if f.ExpectContext != "" {
				ctx := DetectContext(f.Query, f.Position)

				t.Logf("Query: %q (pos %d)", f.Query, f.Position)
				t.Logf("Detected context: %s, prefix: %q", ctx.Type, ctx.Prefix)

				if string(ctx.Type) != f.ExpectContext {
					t.Errorf("Context = %q, want %q", ctx.Type, f.ExpectContext)
				}

				if f.ExpectPrefix != "" || ctx.Prefix != "" {
					if ctx.Prefix != f.ExpectPrefix {
						t.Errorf("Prefix = %q, want %q", ctx.Prefix, f.ExpectPrefix)
					}
				}
			}

			// Test completion generation
			if len(f.ExpectCompletionLabels) > 0 || len(f.ExpectCompletionKinds) > 0 ||
				len(f.ExpectMissingLabels) > 0 || len(f.ExpectFirstCompletions) > 0 {

				compCtx := &CompletionContext{
					Query:           f.Query,
					Position:        f.Position,
					Schema:          s,
					DefaultKeyspace: f.DefaultKeyspace,
				}

				// If we need table context for column completions, we need to set it up
				// This is a bit hacky - we'll modify the detected context
				items := GetCompletions(compCtx)

				t.Logf("Got %d completions", len(items))
				if len(items) > 0 && len(items) <= 10 {
					for _, item := range items {
						t.Logf("  - %s (%s)", item.Label, item.Kind)
					}
				}

				// Check expected labels
				for _, expected := range f.ExpectCompletionLabels {
					found := false
					for _, item := range items {
						if item.Label == expected {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Expected completion %q not found", expected)
					}
				}

				// Check expected kinds
				for _, expectedKind := range f.ExpectCompletionKinds {
					found := false
					for _, item := range items {
						if string(item.Kind) == expectedKind {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Expected completion kind %q not found", expectedKind)
					}
				}

				// Check missing labels
				for _, missing := range f.ExpectMissingLabels {
					for _, item := range items {
						if item.Label == missing {
							t.Errorf("Label %q should not be in completions", missing)
							break
						}
					}
				}

				// Check first completions (order)
				if len(f.ExpectFirstCompletions) > 0 {
					for i, expected := range f.ExpectFirstCompletions {
						if i >= len(items) {
							t.Errorf("Not enough completions, wanted %q at position %d", expected, i)
							continue
						}
						if items[i].Label != expected {
							t.Errorf("Completion[%d] = %q, want %q", i, items[i].Label, expected)
						}
					}
				}
			}
		})
	}
}

// TestContextDetection tests context detection in isolation
func TestContextDetection(t *testing.T) {
	tests := []struct {
		name       string
		query      string
		position   int
		wantType   ContextType
		wantPrefix string
	}{
		{"empty", "", 0, ContextStatementStart, ""},
		{"select keyword", "SELECT", 6, ContextStatementStart, "SELECT"},
		{"after select", "SELECT ", 7, ContextAfterSelect, ""},
		{"after from", "SELECT * FROM ", 14, ContextAfterFrom, ""},
		{"after where", "SELECT * FROM t WHERE ", 22, ContextAfterWhere, ""},
		{"partial keyword", "SEL", 3, ContextStatementStart, "SEL"},
		{"after create", "CREATE ", 7, ContextAfterCreate, ""},
		{"after dot", "SELECT * FROM ks.", 17, ContextAfterDot, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := DetectContext(tt.query, tt.position)
			if ctx.Type != tt.wantType {
				t.Errorf("Type = %v, want %v", ctx.Type, tt.wantType)
			}
			if ctx.Prefix != tt.wantPrefix {
				t.Errorf("Prefix = %q, want %q", ctx.Prefix, tt.wantPrefix)
			}
		})
	}
}

// TestTableContextFromFullQuery tests that table context is extracted from full query
// even when cursor is before FROM clause (for SELECT column completions)
func TestTableContextFromFullQuery(t *testing.T) {
	tests := []struct {
		name         string
		query        string
		position     int
		wantType     ContextType
		wantKeyspace string
		wantTable    string
	}{
		{
			name:         "SELECT cursor before FROM with table",
			query:        "SELECT  FROM users",
			position:     7, // cursor after "SELECT "
			wantType:     ContextAfterSelect,
			wantKeyspace: "",
			wantTable:    "users",
		},
		{
			name:         "SELECT cursor before FROM with keyspace.table",
			query:        "SELECT  FROM myks.users",
			position:     7,
			wantType:     ContextAfterSelect,
			wantKeyspace: "myks",
			wantTable:    "users",
		},
		{
			name:         "SELECT with partial column before FROM",
			query:        "SELECT id,  FROM users",
			position:     11, // cursor after "SELECT id, "
			wantType:     ContextAfterSelect,
			wantKeyspace: "",
			wantTable:    "users",
		},
		{
			name:         "cursor in middle of SELECT columns",
			query:        "SELECT id, name FROM users WHERE x = 1",
			position:     10, // cursor after "SELECT id,"
			wantType:     ContextAfterSelect,
			wantKeyspace: "",
			wantTable:    "users",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := DetectContext(tt.query, tt.position)
			if ctx.Type != tt.wantType {
				t.Errorf("Type = %v, want %v", ctx.Type, tt.wantType)
			}
			if ctx.Keyspace != tt.wantKeyspace {
				t.Errorf("Keyspace = %q, want %q", ctx.Keyspace, tt.wantKeyspace)
			}
			if ctx.Table != tt.wantTable {
				t.Errorf("Table = %q, want %q", ctx.Table, tt.wantTable)
			}
		})
	}
}

// TestCompletionFiltering tests prefix filtering
func TestCompletionFiltering(t *testing.T) {
	items := []CompletionItem{
		{Label: "SELECT", Kind: KindKeyword},
		{Label: "SET", Kind: KindKeyword},
		{Label: "INSERT", Kind: KindKeyword},
	}

	filtered := filterByPrefix(items, "SE")
	if len(filtered) != 2 {
		t.Errorf("Got %d items, want 2", len(filtered))
	}

	filtered = filterByPrefix(items, "sel")
	if len(filtered) != 1 {
		t.Errorf("Got %d items, want 1 (case-insensitive)", len(filtered))
	}
}
