package types

import (
	"strings"
	"testing"
)

func TestErrorString(t *testing.T) {
	err := &Error{
		Line:    1,
		Column:  10,
		Message: "unexpected token",
	}

	str := err.Error()
	if !strings.Contains(str, "line 1:10") {
		t.Errorf("Error string should contain position, got: %s", str)
	}
	if !strings.Contains(str, "unexpected token") {
		t.Errorf("Error string should contain message, got: %s", str)
	}
}

func TestErrorWithSuggestion(t *testing.T) {
	err := &Error{
		Line:       1,
		Column:     10,
		Message:    "unexpected token",
		Suggestion: "Did you mean 'FROM'?",
	}

	str := err.Error()
	if !strings.Contains(str, "suggestion") {
		t.Errorf("Error string should contain suggestion, got: %s", str)
	}
	if !strings.Contains(str, "FROM") {
		t.Errorf("Error string should contain suggestion text, got: %s", str)
	}
}

func TestErrorPosition(t *testing.T) {
	err := &Error{Line: 5, Column: 12}
	pos := err.Position()
	if pos != "5:12" {
		t.Errorf("Position() = %s, want 5:12", pos)
	}
}

func TestErrorsCollection(t *testing.T) {
	var errs Errors
	errs = append(errs, &Error{Line: 1, Column: 0, Message: "error 1"})
	errs = append(errs, &Error{Line: 2, Column: 5, Message: "error 2"})
	errs = append(errs, &Error{Line: 1, Column: 10, Message: "error 3"})

	// Test HasErrors
	if !errs.HasErrors() {
		t.Error("HasErrors() should return true")
	}

	// Test First
	first := errs.First()
	if first.Message != "error 1" {
		t.Errorf("First() returned %q, want 'error 1'", first.Message)
	}

	// Test ByLine
	line1Errors := errs.ByLine(1)
	if len(line1Errors) != 2 {
		t.Errorf("ByLine(1) returned %d errors, want 2", len(line1Errors))
	}

	// Test Error string
	errStr := errs.Error()
	if !strings.Contains(errStr, "3 errors") {
		t.Errorf("Error() should mention count, got: %s", errStr)
	}
}

func TestEmptyErrors(t *testing.T) {
	var errs Errors

	if errs.HasErrors() {
		t.Error("Empty Errors should not have errors")
	}

	if errs.First() != nil {
		t.Error("First() on empty Errors should return nil")
	}

	if errs.Error() != "" {
		t.Error("Error() on empty Errors should return empty string")
	}
}

func TestSingleError(t *testing.T) {
	errs := Errors{
		&Error{Line: 1, Column: 0, Message: "single error"},
	}

	str := errs.Error()
	if strings.Contains(str, "errors:") {
		t.Error("Single error should not show count header")
	}
	if !strings.Contains(str, "single error") {
		t.Errorf("Should contain error message, got: %s", str)
	}
}
