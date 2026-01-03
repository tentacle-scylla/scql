package pipeline

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/tentacle-scylla/scql/codegen"
	"github.com/tentacle-scylla/scql/codegen/grammar"
)

// DebugRules tests the parser rule extraction for debugging purposes.
func DebugRules(dirs *codegen.Dirs) error {
	fmt.Println("=== Debug: Parser Rule Extraction ===")

	content, err := os.ReadFile(filepath.Join(dirs.Grammars, "scylla_Cql.g"))
	if err != nil {
		return fmt.Errorf("failed to read grammar: %w", err)
	}

	rules := grammar.ExtractScyllaParserRules(string(content))

	// Print a few key rules
	testRules := []string{"selectStatement", "pruneMaterializedViewStatement", "insertStatement", "updateStatement", "deleteStatement"}
	for _, name := range testRules {
		if rule, ok := rules[name]; ok {
			fmt.Printf("\n=== %s ===\n%s\n", name, rule)
		} else {
			fmt.Printf("\n=== %s === NOT FOUND\n", name)
		}
	}
	fmt.Printf("\nTotal rules extracted: %d\n", len(rules))

	return nil
}
