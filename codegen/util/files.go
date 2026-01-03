// Package util provides common utility functions for cqlgen.
package util

import (
	"bufio"
	"io"
	"net/http"
	"os"
	"sort"
)

// DownloadFile downloads a file from URL and saves it to the given path.
func DownloadFile(filepath, url string) error {
	resp, err := http.Get(url) //nolint:gosec // URL is from trusted config
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()

	_, err = io.Copy(out, resp.Body)
	return err
}

// WriteLines writes a slice of strings to a file, one per line.
func WriteLines(path string, lines []string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	for _, line := range lines {
		if _, err := f.WriteString(line + "\n"); err != nil {
			return err
		}
	}
	return nil
}

// ReadLines reads all lines from a file.
func ReadLines(path string) []string {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer func() { _ = f.Close() }()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines
}

// ToSet converts a slice of strings to a set (map[string]bool).
func ToSet(items []string) map[string]bool {
	set := make(map[string]bool)
	for _, item := range items {
		set[item] = true
	}
	return set
}

// Unique returns a slice with duplicate strings removed, preserving order.
func Unique(items []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, item := range items {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}
	return result
}

// SortStrings returns a sorted copy of the string slice.
func SortStrings(items []string) []string {
	result := make([]string, len(items))
	copy(result, items)
	sort.Strings(result)
	return result
}
