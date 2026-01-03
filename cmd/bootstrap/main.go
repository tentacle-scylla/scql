// Command bootstrap generates the parser and completion data.
// Use this when the gen/parser Go files are missing.
//
// Usage: go run ./cmd/bootstrap
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pierre-borckmans/scql/codegen"
	"github.com/pierre-borckmans/scql/codegen/pipeline"
)

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	scqlRoot := findScqlRoot(cwd)
	if scqlRoot == "" {
		fmt.Fprintln(os.Stderr, "Error: cannot find scql root directory (looking for gen/)")
		os.Exit(1)
	}

	parserDir := filepath.Join(scqlRoot, "gen", "parser")
	dirs := codegen.NewDirsWithRoot(parserDir, scqlRoot)

	if err := pipeline.All(dirs); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func findScqlRoot(startPath string) string {
	dir := startPath
	for i := 0; i < 10; i++ {
		// Look for gen/ directory or go.mod with scql module
		if fi, err := os.Stat(filepath.Join(dir, "gen")); err == nil && fi.IsDir() {
			return dir
		}
		// Fallback: check for go.mod
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			// Verify it's the scql module by checking for pkg/
			if fi, err := os.Stat(filepath.Join(dir, "pkg")); err == nil && fi.IsDir() {
				return dir
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}
