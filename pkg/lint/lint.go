package lint

import (
	"github.com/tentacle-scylla/scql/pkg/parse"
	"github.com/tentacle-scylla/scql/pkg/types"
)

// Check validates CQL and returns any errors found
func Check(input string) types.Errors {
	result := parse.Parse(input)
	return result.Errors
}

// CheckMultiple validates multiple CQL statements and returns all errors
func CheckMultiple(input string) types.Errors {
	results := parse.Multiple(input)
	var allErrors types.Errors
	for _, r := range results {
		allErrors = append(allErrors, r.Errors...)
	}
	return allErrors
}

// IsValid returns true if the CQL input is syntactically valid
func IsValid(input string) bool {
	return !Check(input).HasErrors()
}

// Result contains detailed lint results for a statement
type Result struct {
	Input   string
	Type    types.StatementType
	Errors  types.Errors
	IsValid bool
}

// Analyze performs detailed analysis on a CQL statement
func Analyze(input string) *Result {
	parseResult := parse.Parse(input)
	return &Result{
		Input:   input,
		Type:    parseResult.Type,
		Errors:  parseResult.Errors,
		IsValid: parseResult.IsValid(),
	}
}

// AnalyzeMultiple performs detailed analysis on multiple CQL statements
func AnalyzeMultiple(input string) []*Result {
	parseResults := parse.Multiple(input)
	var results []*Result
	for _, pr := range parseResults {
		results = append(results, &Result{
			Input:   pr.Input,
			Type:    pr.Type,
			Errors:  pr.Errors,
			IsValid: pr.IsValid(),
		})
	}
	return results
}
