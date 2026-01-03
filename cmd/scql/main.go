package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/tentacle-scylla/scql/pkg/format"
	"github.com/tentacle-scylla/scql/pkg/lint"
	"github.com/tentacle-scylla/scql/pkg/parse"
)

func main() {
	app := &cli.App{
		Name:    "scql",
		Usage:   "CQL parser, linter, and formatter for Cassandra/ScyllaDB",
		Version: "0.1.0",
		Commands: []*cli.Command{
			lintCmd(),
			formatCmd(),
			parseCmd(),
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func lintCmd() *cli.Command {
	return &cli.Command{
		Name:    "lint",
		Aliases: []string{"l", "check"},
		Usage:   "Validate CQL syntax",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "file",
				Aliases: []string{"f"},
				Usage:   "Read CQL from file",
			},
			&cli.BoolFlag{
				Name:    "quiet",
				Aliases: []string{"q"},
				Usage:   "Only output errors, no success message",
			},
		},
		Action: func(c *cli.Context) error {
			input, err := getInput(c)
			if err != nil {
				return err
			}

			results := lint.AnalyzeMultiple(input)
			hasErrors := false
			totalStatements := 0
			validStatements := 0

			for _, r := range results {
				totalStatements++
				if r.IsValid {
					validStatements++
				} else {
					hasErrors = true
					for _, e := range r.Errors {
						fmt.Fprintf(os.Stderr, "%s\n", e.Error())
						if e.Suggestion != "" {
							fmt.Fprintf(os.Stderr, "  suggestion: %s\n", e.Suggestion)
						}
					}
				}
			}

			if !c.Bool("quiet") {
				if hasErrors {
					fmt.Fprintf(os.Stderr, "\n%d/%d statements valid\n", validStatements, totalStatements)
				} else {
					fmt.Printf("OK: %d statements valid\n", totalStatements)
				}
			}

			if hasErrors {
				os.Exit(1)
			}
			return nil
		},
	}
}

func formatCmd() *cli.Command {
	return &cli.Command{
		Name:    "format",
		Aliases: []string{"fmt"},
		Usage:   "Format CQL statements",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "file",
				Aliases: []string{"f"},
				Usage:   "Read CQL from file",
			},
			&cli.BoolFlag{
				Name:    "compact",
				Aliases: []string{"c"},
				Usage:   "Output compact single-line format",
			},
			&cli.BoolFlag{
				Name:  "lowercase",
				Usage: "Use lowercase keywords",
			},
			&cli.StringFlag{
				Name:  "indent",
				Value: "  ",
				Usage: "Indentation string (default: 2 spaces)",
			},
			&cli.BoolFlag{
				Name:    "write",
				Aliases: []string{"w"},
				Usage:   "Write result back to file (requires -f)",
			},
		},
		Action: func(c *cli.Context) error {
			input, err := getInput(c)
			if err != nil {
				return err
			}

			opts := format.DefaultOptions()
			if c.Bool("compact") {
				opts.Style = format.Compact
			}
			if c.Bool("lowercase") {
				opts.UppercaseKeywords = false
			}
			opts.IndentString = c.String("indent")

			results := parse.Multiple(input)
			var outputs []string

			for _, r := range results {
				if r.HasErrors() {
					for _, e := range r.Errors {
						fmt.Fprintf(os.Stderr, "%s\n", e.Error())
					}
					return fmt.Errorf("cannot format invalid CQL")
				}
				outputs = append(outputs, format.Format(r, opts))
			}

			output := strings.Join(outputs, "\n\n")

			if c.Bool("write") && c.String("file") != "" {
				return os.WriteFile(c.String("file"), []byte(output+"\n"), 0644)
			}

			fmt.Println(output)
			return nil
		},
	}
}

func parseCmd() *cli.Command {
	return &cli.Command{
		Name:    "parse",
		Aliases: []string{"p"},
		Usage:   "Parse and analyze CQL statements",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "file",
				Aliases: []string{"f"},
				Usage:   "Read CQL from file",
			},
			&cli.BoolFlag{
				Name:  "json",
				Usage: "Output as JSON (not implemented yet)",
			},
		},
		Action: func(c *cli.Context) error {
			input, err := getInput(c)
			if err != nil {
				return err
			}

			results := lint.AnalyzeMultiple(input)

			for i, r := range results {
				fmt.Printf("Statement %d:\n", i+1)
				fmt.Printf("  Type:  %s\n", r.Type)
				fmt.Printf("  Valid: %v\n", r.IsValid)
				if r.Errors.HasErrors() {
					fmt.Printf("  Errors:\n")
					for _, e := range r.Errors {
						fmt.Printf("    - %s\n", e.Error())
					}
				}
				if i < len(results)-1 {
					fmt.Println()
				}
			}

			return nil
		},
	}
}

func getInput(c *cli.Context) (string, error) {
	// Check for file flag
	if file := c.String("file"); file != "" {
		data, err := os.ReadFile(file)
		if err != nil {
			return "", fmt.Errorf("reading file: %w", err)
		}
		return string(data), nil
	}

	// Check for positional argument
	if c.NArg() > 0 {
		return strings.Join(c.Args().Slice(), " "), nil
	}

	// Check for stdin
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", fmt.Errorf("reading stdin: %w", err)
		}
		return string(data), nil
	}

	// Interactive mode - read until empty line or EOF
	fmt.Fprintln(os.Stderr, "Enter CQL (empty line or Ctrl+D to finish):")
	var lines []string
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			break
		}
		lines = append(lines, line)
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("reading input: %w", err)
	}

	return strings.Join(lines, "\n"), nil
}
