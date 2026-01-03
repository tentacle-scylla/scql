package pipeline

import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/tentacle-scylla/scql/codegen"
	"github.com/tentacle-scylla/scql/codegen/util"
)

// Diff shows differences between ANTLR4 and ScyllaDB grammars.
func Diff(dirs *codegen.Dirs) error {
	fmt.Println("=== Grammar Differences ===")

	antlr4Keywords := util.ReadLines(filepath.Join(dirs.Analysis, "antlr4_keywords.txt"))
	scyllaKeywords := util.ReadLines(filepath.Join(dirs.Analysis, "scylla_keywords.txt"))

	antlr4Set := util.ToSet(antlr4Keywords)
	scyllaSet := util.ToSet(scyllaKeywords)

	// Keywords in ScyllaDB but not in ANTLR4
	fmt.Println("\n--- Keywords in ScyllaDB but NOT in ANTLR4 ---")
	var missing []string
	for kw := range scyllaSet {
		if _, ok := antlr4Set[kw]; !ok {
			missing = append(missing, kw)
		}
	}
	sort.Strings(missing)
	for _, kw := range missing {
		fmt.Printf("  + %s\n", kw)
	}

	// Keywords in ANTLR4 but not in ScyllaDB (these are fine)
	fmt.Println("\n--- Keywords in ANTLR4 but NOT in ScyllaDB (OK) ---")
	var extra []string
	for kw := range antlr4Set {
		if _, ok := scyllaSet[kw]; !ok {
			extra = append(extra, kw)
		}
	}
	sort.Strings(extra)
	for _, kw := range extra {
		fmt.Printf("  - %s\n", kw)
	}

	// Write missing keywords to file for patching
	_ = util.WriteLines(filepath.Join(dirs.Analysis, "missing_keywords.txt"), missing)
	fmt.Printf("\nMissing keywords written to missing_keywords.txt (%d keywords)\n", len(missing))

	return nil
}
