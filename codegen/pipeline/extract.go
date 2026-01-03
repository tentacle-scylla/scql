package pipeline

import (
	"fmt"
	"path/filepath"

	"github.com/pierre-borckmans/scql/codegen"
	"github.com/pierre-borckmans/scql/codegen/grammar"
	"github.com/pierre-borckmans/scql/codegen/util"
)

// Extract extracts keywords and rules from both grammars.
func Extract(dirs *codegen.Dirs) error {
	fmt.Println("=== Extracting grammar components ===")

	// Extract from ANTLR4 lexer
	antlr4Keywords := grammar.ExtractANTLR4Keywords(filepath.Join(dirs.Grammars, "CqlLexer.g4"))
	fmt.Printf("  ANTLR4 keywords: %d\n", len(antlr4Keywords))
	_ = util.WriteLines(filepath.Join(dirs.Analysis, "antlr4_keywords.txt"), antlr4Keywords)

	// Extract from ScyllaDB grammar
	scyllaKeywords := grammar.ExtractScyllaKeywords(filepath.Join(dirs.Grammars, "scylla_Cql.g"))
	fmt.Printf("  ScyllaDB keywords: %d\n", len(scyllaKeywords))
	_ = util.WriteLines(filepath.Join(dirs.Analysis, "scylla_keywords.txt"), scyllaKeywords)

	// Extract parser rules from ANTLR4
	antlr4Rules := grammar.ExtractANTLR4Rules(filepath.Join(dirs.Grammars, "CqlParser.g4"))
	fmt.Printf("  ANTLR4 parser rules: %d\n", len(antlr4Rules))
	_ = util.WriteLines(filepath.Join(dirs.Analysis, "antlr4_rules.txt"), antlr4Rules)

	// Extract parser rules from ScyllaDB
	scyllaRules := grammar.ExtractScyllaRules(filepath.Join(dirs.Grammars, "scylla_Cql.g"))
	fmt.Printf("  ScyllaDB parser rules: %d\n", len(scyllaRules))
	_ = util.WriteLines(filepath.Join(dirs.Analysis, "scylla_rules.txt"), scyllaRules)

	return nil
}
