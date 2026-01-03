// Package pipeline provides the cqlgen pipeline commands.
package pipeline

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pierre-borckmans/scql/codegen"
	"github.com/pierre-borckmans/scql/codegen/util"
)

const scyllaDBRepoURL = "https://github.com/scylladb/scylladb.git"

// Download downloads all required grammars, ANTLR jar, and copies patches.
func Download(dirs *codegen.Dirs) error {
	fmt.Println("=== Downloading grammars ===")

	// Create directory structure
	for _, dir := range []string{dirs.Grammars, dirs.Patched, dirs.Analysis, dirs.Parser, dirs.Patches} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Download grammars to build/grammars/
	grammarDownloads := []struct {
		url  string
		file string
	}{
		{codegen.ANTLR4LexerURL, "CqlLexer.g4"},
		{codegen.ANTLR4ParserURL, "CqlParser.g4"},
		{codegen.ScyllaGrammarURL, "scylla_Cql.g"},
	}

	for _, d := range grammarDownloads {
		path := filepath.Join(dirs.Grammars, d.file)
		if _, err := os.Stat(path); err == nil {
			fmt.Printf("  ✓ %s (exists)\n", d.file)
			continue
		}

		fmt.Printf("  ↓ %s...", d.file)
		if err := util.DownloadFile(path, d.url); err != nil {
			fmt.Printf(" FAILED: %v\n", err)
			continue
		}
		fmt.Println(" OK")
	}

	// Download ANTLR jar to build/
	jarPath := dirs.ANTLRJarPath()
	if _, err := os.Stat(jarPath); err == nil {
		fmt.Printf("  ✓ %s (exists)\n", codegen.ANTLRJarName)
	} else {
		fmt.Printf("  ↓ %s...", codegen.ANTLRJarName)
		if err := util.DownloadFile(jarPath, codegen.ANTLRJarURL); err != nil {
			fmt.Printf(" FAILED: %v\n", err)
		} else {
			fmt.Println(" OK")
		}
	}

	// Copy patches from cqlgen to output directory
	copyPatches(dirs)

	// Download ScyllaDB test files
	if err := downloadScyllaTests(dirs); err != nil {
		fmt.Printf("  Warning: failed to download ScyllaDB tests: %v\n", err)
	}

	// Sparse clone ScyllaDB source for completion data extraction
	if err := cloneScyllaDBSource(dirs); err != nil {
		fmt.Printf("  Warning: failed to clone ScyllaDB source: %v\n", err)
	}

	return nil
}

// copyPatches copies patch files from codegen/patches to the output build directory.
func copyPatches(dirs *codegen.Dirs) {
	fmt.Println("\n=== Copying patches ===")

	// Find the patches directory relative to this package
	// Try multiple locations: codegen/patches, ./patches, and relative to executable
	var srcPatchesDir string
	candidatePaths := []string{
		"codegen/patches",    // When running from scql root
		"./patches",          // Legacy location
		"../codegen/patches", // When running from a subdirectory
	}

	for _, candidate := range candidatePaths {
		if _, err := os.Stat(candidate); err == nil {
			srcPatchesDir = candidate
			break
		}
	}

	if srcPatchesDir == "" {
		// Last resort: try relative to executable
		execPath, err := os.Executable()
		if err == nil {
			srcPatchesDir = filepath.Join(filepath.Dir(execPath), "patches")
		}
	}

	// List and copy all .json and .txt patch files
	files, err := os.ReadDir(srcPatchesDir)
	if err != nil {
		fmt.Printf("  No patches directory found at %s\n", srcPatchesDir)
		return
	}

	for _, f := range files {
		if f.IsDir() {
			continue
		}
		// Copy .json and .txt patch files
		if !strings.HasSuffix(f.Name(), ".json") && !strings.HasSuffix(f.Name(), ".txt") {
			continue
		}

		srcPath := filepath.Join(srcPatchesDir, f.Name())
		destPath := filepath.Join(dirs.Patches, f.Name())

		content, err := os.ReadFile(srcPath)
		if err != nil {
			fmt.Printf("  ✗ %s (read error: %v)\n", f.Name(), err)
			continue
		}

		if err := os.WriteFile(destPath, content, 0644); err != nil {
			fmt.Printf("  ✗ %s (write error: %v)\n", f.Name(), err)
			continue
		}

		fmt.Printf("  ✓ %s\n", f.Name())
	}
}

// githubFile represents a file entry from the GitHub API.
type githubFile struct {
	Name        string `json:"name"`
	DownloadURL string `json:"download_url"`
	Type        string `json:"type"`
}

// downloadScyllaTests downloads CQL test files from ScyllaDB's GitHub repository.
func downloadScyllaTests(dirs *codegen.Dirs) error {
	fmt.Println("\n=== Downloading ScyllaDB test files ===")

	// Create the scylladb tests directory
	if err := os.MkdirAll(dirs.ScyllaTests, 0755); err != nil {
		return fmt.Errorf("failed to create scylladb tests directory: %w", err)
	}

	// Fetch the directory listing from GitHub API
	resp, err := http.Get(codegen.ScyllaTestsAPIURL)
	if err != nil {
		return fmt.Errorf("failed to fetch test file list: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	var files []githubFile
	if err := json.Unmarshal(body, &files); err != nil {
		return fmt.Errorf("failed to parse GitHub response: %w", err)
	}

	// Download each .cql file
	downloaded := 0
	skipped := 0
	for _, f := range files {
		if f.Type != "file" || !strings.HasSuffix(f.Name, ".cql") {
			continue
		}

		destPath := filepath.Join(dirs.ScyllaTests, f.Name)

		// Skip if file already exists
		if _, err := os.Stat(destPath); err == nil {
			skipped++
			continue
		}

		// Download the file
		fileURL := codegen.ScyllaTestsRawURL + "/" + f.Name
		if err := util.DownloadFile(destPath, fileURL); err != nil {
			fmt.Printf("  ✗ %s: %v\n", f.Name, err)
			continue
		}
		downloaded++
	}

	if downloaded > 0 {
		fmt.Printf("  ↓ Downloaded %d new test files\n", downloaded)
	}
	if skipped > 0 {
		fmt.Printf("  ✓ %d test files already present\n", skipped)
	}

	return nil
}

// cloneScyllaDBSource performs a sparse clone of ScyllaDB to get only the
// directories needed for completion data extraction (cql3/, types/).
func cloneScyllaDBSource(dirs *codegen.Dirs) error {
	fmt.Println("\n=== Cloning ScyllaDB source (sparse) ===")

	scyllaDir := dirs.ScyllaDB

	// Check if already cloned
	if _, err := os.Stat(filepath.Join(scyllaDir, ".git")); err == nil {
		// Already cloned, do a pull to update
		fmt.Println("  ✓ ScyllaDB source already present, updating...")
		cmd := exec.Command("git", "pull", "--depth", "1")
		cmd.Dir = scyllaDir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			// Pull failed, but we have existing data, so continue
			fmt.Printf("  Warning: git pull failed: %v (using existing data)\n", err)
		}
		return nil
	}

	// Create the directory
	if err := os.MkdirAll(scyllaDir, 0755); err != nil {
		return fmt.Errorf("failed to create scylladb directory: %w", err)
	}

	fmt.Println("  Initializing sparse checkout...")

	// Initialize empty repo
	cmd := exec.Command("git", "init")
	cmd.Dir = scyllaDir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git init failed: %w", err)
	}

	// Add remote
	cmd = exec.Command("git", "remote", "add", "origin", scyllaDBRepoURL)
	cmd.Dir = scyllaDir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git remote add failed: %w", err)
	}

	// Enable sparse checkout
	cmd = exec.Command("git", "config", "core.sparseCheckout", "true")
	cmd.Dir = scyllaDir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git config sparse checkout failed: %w", err)
	}

	// Configure sparse checkout paths
	sparseFile := filepath.Join(scyllaDir, ".git", "info", "sparse-checkout")
	sparsePaths := []string{
		"cql3/",   // Grammar and function definitions
		"types/",  // Type definitions
	}
	if err := os.WriteFile(sparseFile, []byte(strings.Join(sparsePaths, "\n")+"\n"), 0644); err != nil {
		return fmt.Errorf("failed to write sparse-checkout file: %w", err)
	}

	// Fetch with depth 1 (shallow clone)
	fmt.Println("  Fetching cql3/ and types/ directories...")
	cmd = exec.Command("git", "fetch", "--depth", "1", "origin", "master")
	cmd.Dir = scyllaDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git fetch failed: %w", err)
	}

	// Checkout
	cmd = exec.Command("git", "checkout", "master")
	cmd.Dir = scyllaDir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git checkout failed: %w", err)
	}

	fmt.Println("  ✓ ScyllaDB source cloned (sparse: cql3/, types/)")
	return nil
}
