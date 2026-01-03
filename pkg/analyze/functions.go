package analyze

import (
	"fmt"
	"strings"

	"github.com/tentacle-scylla/scql/gen/cqldata"
)

// FunctionSignature describes the expected arguments for a built-in function.
type FunctionSignature struct {
	Name      string
	MinArgs   int  // Minimum number of arguments
	MaxArgs   int  // Maximum number of arguments (-1 = unlimited)
	AllowStar bool // Whether * is a valid argument (e.g., count(*))
}

// builtinFunctions maps function names to their signatures.
var builtinFunctions map[string]*FunctionSignature

func init() {
	builtinFunctions = buildFunctionSignatures()
}

// buildFunctionSignatures creates the signature map from generated data.
func buildFunctionSignatures() map[string]*FunctionSignature {
	sigs := make(map[string]*FunctionSignature)

	// Build from generated functions
	for _, f := range cqldata.GenFunctions {
		name := strings.ToLower(f.Name)

		// Check if we already have this function (some have overloads)
		if existing, ok := sigs[name]; ok {
			// Update to allow range if different param counts
			paramCount := len(f.Params)
			if paramCount < existing.MinArgs {
				existing.MinArgs = paramCount
			}
			if paramCount > existing.MaxArgs {
				existing.MaxArgs = paramCount
			}
			continue
		}

		paramCount := len(f.Params)
		sigs[name] = &FunctionSignature{
			Name:    name,
			MinArgs: paramCount,
			MaxArgs: paramCount,
		}
	}

	// Apply special overrides for functions with special behavior
	applySpecialCases(sigs)

	return sigs
}

// applySpecialCases adds overrides for functions that need special handling.
func applySpecialCases(sigs map[string]*FunctionSignature) {
	// count() requires either * or exactly 1 column argument
	// count() with no args is invalid, count(*) or count(col) is valid
	if sig, ok := sigs["count"]; ok {
		sig.MinArgs = 1 // Requires at least 1 arg (or *)
		sig.MaxArgs = 1
		sig.AllowStar = true
	}

	// token() takes 1 or more partition key columns
	if sig, ok := sigs["token"]; ok {
		sig.MinArgs = 1
		sig.MaxArgs = -1 // Unlimited (composite partition key)
	}

	// uuid() takes no arguments
	if _, ok := sigs["uuid"]; !ok {
		sigs["uuid"] = &FunctionSignature{Name: "uuid", MinArgs: 0, MaxArgs: 0}
	} else {
		sigs["uuid"].MinArgs = 0
		sigs["uuid"].MaxArgs = 0
	}

	// now() takes no arguments
	if _, ok := sigs["now"]; !ok {
		sigs["now"] = &FunctionSignature{Name: "now", MinArgs: 0, MaxArgs: 0}
	} else {
		sigs["now"].MinArgs = 0
		sigs["now"].MaxArgs = 0
	}

	// ttl() takes exactly 1 column
	if sig, ok := sigs["ttl"]; ok {
		sig.MinArgs = 1
		sig.MaxArgs = 1
	}

	// writetime() takes exactly 1 column
	if sig, ok := sigs["writetime"]; ok {
		sig.MinArgs = 1
		sig.MaxArgs = 1
	}

	// cast() takes 2 arguments (value, type)
	if sig, ok := sigs["cast"]; ok {
		sig.MinArgs = 2
		sig.MaxArgs = 2
	}

	// Timestamp functions
	if _, ok := sigs["currentdate"]; !ok {
		sigs["currentdate"] = &FunctionSignature{Name: "currentdate", MinArgs: 0, MaxArgs: 0}
	}
	if _, ok := sigs["currenttime"]; !ok {
		sigs["currenttime"] = &FunctionSignature{Name: "currenttime", MinArgs: 0, MaxArgs: 0}
	}
	if _, ok := sigs["currenttimestamp"]; !ok {
		sigs["currenttimestamp"] = &FunctionSignature{Name: "currenttimestamp", MinArgs: 0, MaxArgs: 0}
	}
	if _, ok := sigs["currenttimeuuid"]; !ok {
		sigs["currenttimeuuid"] = &FunctionSignature{Name: "currenttimeuuid", MinArgs: 0, MaxArgs: 0}
	}
}

// ValidateFunctionCalls validates function calls against known signatures.
// Returns errors for invalid argument counts.
func ValidateFunctionCalls(calls []*FunctionCall) []*SchemaError {
	var errors []*SchemaError

	for _, call := range calls {
		if err := validateFunctionCall(call); err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}

// validateFunctionCall validates a single function call.
func validateFunctionCall(call *FunctionCall) *SchemaError {
	sig, ok := builtinFunctions[call.Name]
	if !ok {
		// Unknown function - could be a UDF, don't error
		return nil
	}

	// Handle star argument (e.g., count(*))
	if call.HasStar {
		if !sig.AllowStar {
			return &SchemaError{
				Type:     ErrFunctionArgCount,
				Message:  fmt.Sprintf("%s() does not accept * as argument", call.Name),
				Object:   call.Name,
				Position: call.Position,
			}
		}
		// Star counts as a valid argument for functions that allow it
		return nil
	}

	// Check argument count
	if sig.MaxArgs == -1 {
		// Unlimited args, just check minimum
		if call.ArgCount < sig.MinArgs {
			return &SchemaError{
				Type:     ErrFunctionArgCount,
				Message:  fmt.Sprintf("%s() requires at least %d argument(s), got %d", call.Name, sig.MinArgs, call.ArgCount),
				Object:   call.Name,
				Position: call.Position,
			}
		}
	} else if sig.MinArgs == sig.MaxArgs {
		// Exact number required
		if call.ArgCount != sig.MinArgs {
			return &SchemaError{
				Type:     ErrFunctionArgCount,
				Message:  fmt.Sprintf("%s() takes %d argument(s), got %d", call.Name, sig.MinArgs, call.ArgCount),
				Object:   call.Name,
				Position: call.Position,
			}
		}
	} else {
		// Range allowed
		if call.ArgCount < sig.MinArgs || call.ArgCount > sig.MaxArgs {
			return &SchemaError{
				Type:     ErrFunctionArgCountRange,
				Message:  fmt.Sprintf("%s() takes %d-%d argument(s), got %d", call.Name, sig.MinArgs, sig.MaxArgs, call.ArgCount),
				Object:   call.Name,
				Position: call.Position,
			}
		}
	}

	return nil
}

// GetFunctionSignature returns the signature for a built-in function, or nil if unknown.
func GetFunctionSignature(name string) *FunctionSignature {
	return builtinFunctions[strings.ToLower(name)]
}
