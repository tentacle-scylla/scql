package tests

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/antlr4-go/antlr/v4"
	parser "github.com/pierre-borckmans/scql/gen/parser"
)

// TestParseQueries runs all .cql test files through the parser
func TestParseQueries(t *testing.T) {
	testsDir := "queries"

	// Find all .cql files in queries/ and queries/scylladb/
	files, err := filepath.Glob(filepath.Join(testsDir, "*.cql"))
	if err != nil {
		t.Fatalf("Failed to find test files: %v", err)
	}

	scyllaFiles, err := filepath.Glob(filepath.Join(testsDir, "scylladb", "*.cql"))
	if err != nil {
		t.Fatalf("Failed to find ScyllaDB test files: %v", err)
	}
	files = append(files, scyllaFiles...)

	// Include extra features tests (ScyllaDB extensions)
	extraFiles, err := filepath.Glob(filepath.Join(testsDir, "extra_features", "*.cql"))
	if err != nil {
		t.Fatalf("Failed to find extra features test files: %v", err)
	}
	files = append(files, extraFiles...)

	if len(files) == 0 {
		t.Fatal("No .cql test files found")
	}

	var totalQueries, passed, failed int

	for _, file := range files {
		queries := loadQueries(file)
		t.Logf("Testing %s (%d queries)", filepath.Base(file), len(queries))

		for _, q := range queries {
			totalQueries++
			if err := parseQuery(q); err != nil {
				failed++
				t.Errorf("FAIL: %s\n      Query: %s", err, truncate(q, 80))
			} else {
				passed++
			}
		}
	}

	coverage := float64(passed) / float64(totalQueries) * 100
	t.Logf("\n=== Summary ===")
	t.Logf("Total: %d, Passed: %d, Failed: %d (%.1f%% coverage)", totalQueries, passed, failed, coverage)

	if failed > 0 {
		t.Errorf("%d queries failed to parse", failed)
	}
}

// loadQueries reads CQL queries from a file, one query per semicolon
// Skips queries in "error checks" sections (used for testing invalid syntax)
func loadQueries(path string) []string {
	file, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer func() { _ = file.Close() }()

	var queries []string
	var current strings.Builder
	inErrorSection := false

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Check for error section markers
		if strings.HasPrefix(line, "--") {
			lowerLine := strings.ToLower(line)
			if strings.Contains(lowerLine, "error check") {
				inErrorSection = true
			}
			continue
		}

		// Skip empty lines
		if line == "" {
			continue
		}

		// Skip queries in error sections (these are intentionally invalid)
		if inErrorSection {
			if strings.HasSuffix(line, ";") {
				// Skip this query, it's an error test case
				continue
			}
			continue
		}

		current.WriteString(line)
		current.WriteString(" ")

		// Complete query on semicolon
		if strings.HasSuffix(line, ";") {
			q := strings.TrimSpace(current.String())
			if q != "" {
				queries = append(queries, q)
			}
			current.Reset()
		}
	}

	return queries
}

// errorListener captures parse errors
type errorListener struct {
	*antlr.DefaultErrorListener
	errors []string
}

func (l *errorListener) SyntaxError(_ antlr.Recognizer, _ any, line, column int, msg string, _ antlr.RecognitionException) {
	l.errors = append(l.errors, fmt.Sprintf("line %d:%d %s", line, column, msg))
}

// parseQuery attempts to parse a CQL query and returns any error
func parseQuery(query string) error {
	input := antlr.NewInputStream(query)
	lexer := parser.NewCqlLexer(input)
	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	p := parser.NewCqlParser(stream)

	errListener := &errorListener{}
	lexer.RemoveErrorListeners()
	lexer.AddErrorListener(errListener)
	p.RemoveErrorListeners()
	p.AddErrorListener(errListener)

	p.Root()

	if len(errListener.errors) > 0 {
		return fmt.Errorf("%s", errListener.errors[0])
	}
	return nil
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
