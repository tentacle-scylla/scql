package types

import (
	"fmt"
	"strings"
)

// Error represents a parsing or validation error with position information
type Error struct {
	Line            int    // 1-based line number
	Column          int    // 0-based column number
	Message         string // Raw ANTLR error message (kept for debugging)
	FriendlyMessage string // User-friendly error message (shown in UI)
	Query           string // The original query (or portion) that caused the error
	Suggestion      string // Optional suggestion for fixing the error
}

// Error implements the error interface
func (e *Error) Error() string {
	msg := e.FriendlyMessage
	if msg == "" {
		msg = e.Message
	}
	if e.Suggestion != "" {
		return fmt.Sprintf("line %d:%d: %s (suggestion: %s)", e.Line, e.Column, msg, e.Suggestion)
	}
	return fmt.Sprintf("line %d:%d: %s", e.Line, e.Column, msg)
}

// DisplayMessage returns the best message to show to users
// (FriendlyMessage if available, otherwise raw Message)
func (e *Error) DisplayMessage() string {
	if e.FriendlyMessage != "" {
		return e.FriendlyMessage
	}
	return e.Message
}

// Position returns a string representation of the error position
func (e *Error) Position() string {
	return fmt.Sprintf("%d:%d", e.Line, e.Column)
}

// Errors is a collection of Error pointers
type Errors []*Error

// Error implements the error interface for the collection
func (e Errors) Error() string {
	if len(e) == 0 {
		return ""
	}
	if len(e) == 1 {
		return e[0].Error()
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("%d errors:\n", len(e)))
	for i, err := range e {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString("  - ")
		b.WriteString(err.Error())
	}
	return b.String()
}

// HasErrors returns true if there are any errors
func (e Errors) HasErrors() bool {
	return len(e) > 0
}

// First returns the first error or nil if empty
func (e Errors) First() *Error {
	if len(e) == 0 {
		return nil
	}
	return e[0]
}

// ByLine returns all errors at a specific line
func (e Errors) ByLine(line int) Errors {
	var result Errors
	for _, err := range e {
		if err.Line == line {
			result = append(result, err)
		}
	}
	return result
}
