package pipeline

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pierre-borckmans/scql/codegen"
)

// Generate generates the Go parser from the patched grammar.
func Generate(dirs *codegen.Dirs) error {
	fmt.Println("=== Generating Go Parser ===")

	// Convert to absolute path to avoid issues
	absBase, err := filepath.Abs(dirs.Base)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Recalculate paths with absolute base
	absDirs := codegen.NewDirs(absBase)

	jarPath := absDirs.ANTLRJarPath()
	lexerPath := filepath.Join(absDirs.Patched, "CqlLexer.g4")
	parserPath := filepath.Join(absDirs.Patched, "CqlParser.g4")
	parserOutDir := absDirs.Parser

	// Verify patched grammars exist
	if _, err := os.Stat(lexerPath); os.IsNotExist(err) {
		return fmt.Errorf("patched lexer not found - run 'patch' command first")
	}

	// Remove old parser files but preserve important files/dirs
	entries, _ := os.ReadDir(parserOutDir)
	for _, entry := range entries {
		name := entry.Name()
		// Skip: tests/, build/, .gitignore, go.mod, go.sum
		if name == "tests" || name == "build" || name == ".gitignore" || name == "go.mod" || name == "go.sum" {
			continue
		}
		_ = os.RemoveAll(filepath.Join(parserOutDir, name))
	}
	_ = os.MkdirAll(parserOutDir, 0755)

	// Run ANTLR4
	cmd := exec.Command("java", "-jar", jarPath,
		"-Dlanguage=Go", "-package", "generated_parser", "-o", parserOutDir,
		lexerPath, parserPath)
	cmd.Dir = absBase
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ANTLR4 failed: %w", err)
	}

	// Verify build from generated_parser module
	cmd = exec.Command("go", "build", "./...")
	cmd.Dir = parserOutDir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	fmt.Println("✓ Parser generated and compiled successfully")
	return nil
}
