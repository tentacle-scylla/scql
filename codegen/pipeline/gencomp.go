package pipeline

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pierre-borckmans/scql/codegen"
	"github.com/pierre-borckmans/scql/codegen/cqlextract"
)

// GenerateComplete generates the completion data files from ScyllaDB source.
// This extracts keywords, functions, and types from the ScyllaDB grammar and source.
func GenerateComplete(dirs *codegen.Dirs) error {
	fmt.Println("=== Generating Completion Data ===")

	// Check if ScyllaDB source is available
	if dirs.ScyllaDB == "" || dirs.GenCqlData == "" {
		fmt.Println("  Skipping: ScyllaDB source or GenCqlData path not configured")
		return nil
	}

	scyllaDir := dirs.ScyllaDB
	grammarPath := filepath.Join(scyllaDir, "cql3", "Cql.g")

	// Verify ScyllaDB source exists
	if _, err := os.Stat(grammarPath); os.IsNotExist(err) {
		fmt.Println("  Skipping: ScyllaDB grammar not found (run 'download' first)")
		return nil
	}

	// Ensure output directory exists
	if err := os.MkdirAll(dirs.GenCqlData, 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	// Parse grammar for keywords
	fmt.Println("  Parsing grammar for keywords...")
	grammarData, err := cqlextract.ParseGrammar(grammarPath)
	if err != nil {
		return fmt.Errorf("parsing grammar: %w", err)
	}
	fmt.Printf("    Found %d keywords\n", len(grammarData.AllKeywords))

	// Parse functions
	fmt.Println("  Parsing function definitions...")
	functionsDir := filepath.Join(scyllaDir, "cql3", "functions")
	functions, err := cqlextract.ParseFunctions(functionsDir)
	if err != nil {
		return fmt.Errorf("parsing functions: %w", err)
	}
	fmt.Printf("    Found %d functions\n", len(functions))

	// Parse types
	fmt.Println("  Parsing type definitions...")
	typesPath := filepath.Join(scyllaDir, "types", "types.cc")
	typeData, err := cqlextract.ParseTypes(typesPath)
	if err != nil {
		return fmt.Errorf("parsing types: %w", err)
	}
	fmt.Printf("    Found %d types\n", len(typeData.Types))

	// Generate output files
	fmt.Println("  Generating Go files...")

	if err := cqlextract.GenerateKeywordsFile(grammarData, dirs.GenCqlData); err != nil {
		return fmt.Errorf("generating keywords file: %w", err)
	}
	fmt.Println("    ✓ gen_keywords.go")

	if err := cqlextract.GenerateFunctionsFile(functions, dirs.GenCqlData); err != nil {
		return fmt.Errorf("generating functions file: %w", err)
	}
	fmt.Println("    ✓ gen_functions.go")

	if err := cqlextract.GenerateTypesFile(typeData, dirs.GenCqlData); err != nil {
		return fmt.Errorf("generating types file: %w", err)
	}
	fmt.Println("    ✓ gen_types.go")

	fmt.Println("✓ Completion data generated successfully")
	return nil
}
