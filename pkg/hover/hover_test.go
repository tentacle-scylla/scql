package hover

import (
	"os"
	"strings"
	"testing"

	"github.com/pierre-borckmans/scql/pkg/schema"
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

// HoverFixture represents a single test case
type HoverFixture struct {
	Name     string `yaml:"name"`
	Query    string `yaml:"query"`
	Position int    `yaml:"position"`
	Comment  string `yaml:"comment,omitempty"`

	// Schema context
	SchemaRef       string `yaml:"schemaRef,omitempty"`
	DefaultKeyspace string `yaml:"defaultKeyspace,omitempty"`

	// Token expectations
	ExpectToken      string `yaml:"expectToken,omitempty"`
	ExpectTokenStart int    `yaml:"expectTokenStart,omitempty"`
	ExpectTokenEnd   int    `yaml:"expectTokenEnd,omitempty"`

	// Hover expectations
	ExpectKind            string `yaml:"expectKind,omitempty"`
	ExpectName            string `yaml:"expectName,omitempty"`
	ExpectContentContains string `yaml:"expectContentContains,omitempty"`
	ExpectNoHover         bool   `yaml:"expectNoHover,omitempty"`
}

// FixtureFile represents the entire fixture file structure
type FixtureFile struct {
	Schemas map[string]FixtureSchema `yaml:"schemas"`
	Tests   []HoverFixture           `yaml:"tests"`
}

func loadFixtures(t *testing.T) (*FixtureFile, error) {
	data, err := os.ReadFile("testdata/hover_fixtures.yaml")
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
		}
	}
	return s
}

func TestHoverFixtures(t *testing.T) {
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

			// Test token finding
			if f.ExpectToken != "" || f.ExpectTokenStart != 0 || f.ExpectTokenEnd != 0 {
				token := FindTokenAtPosition(f.Query, f.Position)

				t.Logf("Query: %q (pos %d)", f.Query, f.Position)
				if token != nil {
					t.Logf("Found token: %q [%d:%d] type=%s", token.Text, token.Start, token.End, token.Type)
				} else {
					t.Logf("No token found")
				}

				if f.ExpectToken == "" {
					if token != nil && token.Text != "" {
						t.Errorf("Expected no token, got %q", token.Text)
					}
				} else {
					if token == nil {
						t.Errorf("Expected token %q, got nil", f.ExpectToken)
					} else {
						if token.Text != f.ExpectToken {
							t.Errorf("Token = %q, want %q", token.Text, f.ExpectToken)
						}
						if f.ExpectTokenStart != 0 && token.Start != f.ExpectTokenStart {
							t.Errorf("TokenStart = %d, want %d", token.Start, f.ExpectTokenStart)
						}
						if f.ExpectTokenEnd != 0 && token.End != f.ExpectTokenEnd {
							t.Errorf("TokenEnd = %d, want %d", token.End, f.ExpectTokenEnd)
						}
					}
				}
			}

			// Test hover info
			if f.ExpectKind != "" || f.ExpectNoHover {
				ctx := &HoverContext{
					Query:           f.Query,
					Position:        f.Position,
					Schema:          s,
					DefaultKeyspace: f.DefaultKeyspace,
				}

				info := GetHoverInfo(ctx)

				t.Logf("Query: %q (pos %d)", f.Query, f.Position)
				if info != nil {
					t.Logf("Hover: kind=%s, name=%s", info.Kind, info.Name)
					t.Logf("Content: %s", info.Content)
				} else {
					t.Logf("No hover info")
				}

				if f.ExpectNoHover {
					if info != nil {
						t.Errorf("Expected no hover, got kind=%s name=%s", info.Kind, info.Name)
					}
				} else {
					if info == nil {
						t.Errorf("Expected hover with kind=%q, got nil", f.ExpectKind)
					} else {
						if f.ExpectKind != "" && string(info.Kind) != f.ExpectKind {
							t.Errorf("Kind = %q, want %q", info.Kind, f.ExpectKind)
						}
						if f.ExpectName != "" && !strings.EqualFold(info.Name, f.ExpectName) {
							t.Errorf("Name = %q, want %q", info.Name, f.ExpectName)
						}
						if f.ExpectContentContains != "" && !strings.Contains(info.Content, f.ExpectContentContains) {
							t.Errorf("Content %q does not contain %q", info.Content, f.ExpectContentContains)
						}
					}
				}
			}
		})
	}
}

// TestTokenFinding tests token finding in isolation
func TestTokenFinding(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		position int
		wantText string
		wantType TokenType
	}{
		{"keyword", "SELECT", 3, "SELECT", TokenKeyword},
		{"identifier", "users", 2, "users", TokenIdentifier},
		{"function", "now()", 1, "now", TokenFunction},
		{"at start", "SELECT", 0, "SELECT", TokenKeyword},
		{"at end", "SELECT", 6, "SELECT", TokenKeyword},
		{"in whitespace", "a b", 1, "", ""},
		{"empty", "", 0, "", ""},
		{"underscore", "user_id", 4, "user_id", TokenIdentifier},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := FindTokenAtPosition(tt.query, tt.position)
			if tt.wantText == "" {
				if token != nil && token.Text != "" {
					t.Errorf("Token = %q, want empty", token.Text)
				}
			} else {
				if token == nil {
					t.Errorf("Token = nil, want %q", tt.wantText)
				} else {
					if token.Text != tt.wantText {
						t.Errorf("Token.Text = %q, want %q", token.Text, tt.wantText)
					}
					if token.Type != tt.wantType {
						t.Errorf("Token.Type = %v, want %v", token.Type, tt.wantType)
					}
				}
			}
		})
	}
}

// TestKeywordHover tests hover for keywords
func TestKeywordHover(t *testing.T) {
	keywords := []string{"SELECT", "INSERT", "UPDATE", "DELETE", "CREATE", "DROP", "ALTER"}

	for _, kw := range keywords {
		t.Run(kw, func(t *testing.T) {
			info := GetKeywordInfo(kw)
			if info == nil {
				t.Errorf("No info for keyword %s", kw)
			} else {
				if info.Description == "" {
					t.Errorf("Empty description for keyword %s", kw)
				}
			}
		})
	}
}

// TestFunctionHover tests hover for functions
func TestFunctionHover(t *testing.T) {
	functions := []string{"now", "uuid", "token", "count", "sum", "avg", "min", "max", "writetime", "ttl"}

	for _, fn := range functions {
		t.Run(fn, func(t *testing.T) {
			info := GetFunctionInfo(fn)
			if info == nil {
				t.Errorf("No info for function %s", fn)
			} else {
				if info.Signature == "" {
					t.Errorf("Empty signature for function %s", fn)
				}
				if info.ReturnType == "" {
					t.Errorf("Empty return type for function %s", fn)
				}
			}
		})
	}
}

// TestTypeHover tests hover for types
func TestTypeHover(t *testing.T) {
	types := []string{"int", "bigint", "text", "uuid", "timeuuid", "timestamp", "boolean", "blob"}

	for _, ty := range types {
		t.Run(ty, func(t *testing.T) {
			info := GetTypeInfo(ty)
			if info == nil {
				t.Errorf("No info for type %s", ty)
			} else {
				if info.Description == "" {
					t.Errorf("Empty description for type %s", ty)
				}
			}
		})
	}
}

// TestSchemaAwareHover tests hover with schema context
func TestSchemaAwareHover(t *testing.T) {
	// Build test schema
	s := schema.NewSchema()
	ks := s.AddKeyspace("myapp")
	tbl := ks.AddTable("users")
	tbl.AddColumn("id", "uuid")
	tbl.AddColumn("name", "text")
	tbl.SetPartitionKey("id")

	tests := []struct {
		name     string
		query    string
		position int
		wantKind HoverKind
		wantName string
	}{
		{"column hover", "SELECT id FROM myapp.users", 8, HoverColumn, "id"},
		{"table hover", "SELECT * FROM users", 16, HoverTable, "users"},
		{"keyspace hover", "USE myapp", 6, HoverKeyspace, "myapp"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &HoverContext{
				Query:           tt.query,
				Position:        tt.position,
				Schema:          s,
				DefaultKeyspace: "myapp",
			}

			info := GetHoverInfo(ctx)
			if info == nil {
				t.Fatalf("Expected hover info, got nil")
			}

			if info.Kind != tt.wantKind {
				t.Errorf("Kind = %v, want %v", info.Kind, tt.wantKind)
			}
			if !strings.EqualFold(info.Name, tt.wantName) {
				t.Errorf("Name = %v, want %v", info.Name, tt.wantName)
			}
		})
	}
}
