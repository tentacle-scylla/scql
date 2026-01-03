package pipeline

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/tentacle-scylla/scql/codegen"
)

// Test runs the test coverage for the generated parser.
func Test(dirs *codegen.Dirs) error {
	fmt.Println("=== Running Test Coverage ===")

	absDir, _ := filepath.Abs(dirs.Base)

	// Run go test
	cmd := exec.Command("go", "test", "-v", "./...")
	cmd.Dir = absDir
	cmd.Env = append(os.Environ(), "GOTOOLCHAIN=local")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Don't fail on test failures, just show results
	_ = cmd.Run()
	return nil
}
