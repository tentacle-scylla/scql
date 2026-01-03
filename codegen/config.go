// Package codegen provides configuration and directory helpers for code generation.
package codegen

import "path/filepath"

// URLs for downloading grammars and tools.
const (
	ANTLR4LexerURL   = "https://raw.githubusercontent.com/antlr/grammars-v4/master/cql3/CqlLexer.g4"
	ANTLR4ParserURL  = "https://raw.githubusercontent.com/antlr/grammars-v4/master/cql3/CqlParser.g4"
	ScyllaGrammarURL = "https://raw.githubusercontent.com/scylladb/scylladb/master/cql3/Cql.g"
	ANTLRJarURL      = "https://www.antlr.org/download/antlr-4.13.2-complete.jar"
	ANTLRJarName     = "antlr-4.13.2-complete.jar"

	// ScyllaDB test files API URL
	ScyllaTestsAPIURL = "https://api.github.com/repos/scylladb/scylladb/contents/test/cql"
	ScyllaTestsRawURL = "https://raw.githubusercontent.com/scylladb/scylladb/master/test/cql"
)

// Dirs holds all directory paths for the pipeline.
type Dirs struct {
	Base        string
	Build       string
	Grammars    string
	Patched     string
	Analysis    string
	Parser     string
	GenCqlData string
	Patches    string
	Tests       string
	ScyllaTests string
	ScyllaDB    string // Path to local ScyllaDB source (for completion data extraction)
}

// NewDirs creates a Dirs structure for the given base output directory.
// scqlRoot is the root of the scql library (parent of gen/).
func NewDirs(base string) *Dirs {
	build := filepath.Join(base, "build")
	tests := filepath.Join(base, "tests", "queries")
	return &Dirs{
		Base:        base,
		Build:       build,
		Grammars:    filepath.Join(build, "grammars"),
		Patched:     filepath.Join(build, "patched"),
		Analysis:    filepath.Join(build, "analysis"),
		Parser:     base, // Parser files go in gen/parser root
		GenCqlData: "",   // Set by NewDirsWithRoot
		Patches:     filepath.Join(build, "patches"),
		Tests:       tests,
		ScyllaTests: filepath.Join(tests, "scylladb"),
		ScyllaDB:    filepath.Join(build, "scylladb"),
	}
}

// NewDirsWithRoot creates a Dirs structure with both parser base and scql root.
func NewDirsWithRoot(parserBase, scqlRoot string) *Dirs {
	dirs := NewDirs(parserBase)
	dirs.GenCqlData = filepath.Join(scqlRoot, "gen", "cqldata")
	return dirs
}

// ANTLRJarPath returns the full path to the ANTLR jar file.
func (d *Dirs) ANTLRJarPath() string {
	return filepath.Join(d.Build, ANTLRJarName)
}

// LexerPath returns the path to a lexer file in the specified directory.
func (d *Dirs) LexerPath(dir string) string {
	return filepath.Join(dir, "CqlLexer.g4")
}

// ParserPath returns the path to a parser file in the specified directory.
func (d *Dirs) ParserPath(dir string) string {
	return filepath.Join(dir, "CqlParser.g4")
}
