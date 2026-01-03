package complete

import (
	"strings"

	"github.com/tentacle-scylla/scql/gen/cqldata"
)

// GetFunctionsForType returns functions that produce the given type.
// This enables type-aware completions in WHERE clauses.
func GetFunctionsForType(targetType string) []CompletionItem {
	targetType = strings.ToLower(targetType)

	// Get compatible types
	compatTypes := []string{targetType}
	if compat, ok := cqldata.GenTypeCompatibility[targetType]; ok {
		compatTypes = append(compatTypes, compat...)
	}

	// Build completion items for functions returning compatible types
	var items []CompletionItem
	seen := make(map[string]bool)

	for _, t := range compatTypes {
		if funcs, ok := cqldata.GenFunctionsByReturnType[t]; ok {
			for _, fname := range funcs {
				if seen[fname] {
					continue
				}
				seen[fname] = true

				// Find function details
				for _, f := range cqldata.GenFunctions {
					if f.Name == fname {
						items = append(items, CompletionItem{
							Label:        fname + "()",
							Kind:         KindFunction,
							Detail:       "Returns " + f.ReturnType,
							InsertText:   fname + "()",
							SortPriority: 50,
						})
						break
					}
				}
			}
		}
	}

	return items
}

// GetValueFunctionsForContext returns functions that can provide values,
// optionally filtered by column type.
func GetValueFunctionsForContext(columnType string) []CompletionItem {
	// If column type is known, get type-aware suggestions
	if columnType != "" {
		return GetFunctionsForType(columnType)
	}

	// Otherwise, return common value functions
	valueFuncs := []string{
		"now", "uuid", "currentTimestamp", "currentDate", "currentTime",
		"currentTimeUUID", "toTimestamp", "toDate",
	}

	var items []CompletionItem
	for _, fname := range valueFuncs {
		for _, f := range cqldata.GenFunctions {
			if f.Name == fname {
				items = append(items, CompletionItem{
					Label:        fname + "()",
					Kind:         KindFunction,
					Detail:       "Returns " + f.ReturnType,
					InsertText:   fname + "()",
					SortPriority: 50,
				})
				break
			}
		}
	}

	return items
}

// IsKeyword checks if a word is a CQL keyword.
func IsKeyword(word string) bool {
	word = strings.ToUpper(word)
	for _, kw := range cqldata.GenAllKeywords {
		if word == kw {
			return true
		}
	}
	return false
}

// IsTypeKeyword checks if a word is a CQL type keyword.
func IsTypeKeyword(word string) bool {
	word = strings.ToUpper(word)
	for _, kw := range cqldata.GenTypeKeywords {
		if word == kw {
			return true
		}
	}
	return false
}

// IsUnreservedKeyword checks if a keyword can be used as an identifier.
func IsUnreservedKeyword(word string) bool {
	word = strings.ToUpper(word)
	for _, kw := range cqldata.GenUnreservedKeywords {
		if word == kw {
			return true
		}
	}
	return false
}

// GetTypeKind returns the kind/category of a CQL type.
func GetTypeKind(typeName string) string {
	typeName = strings.ToLower(typeName)
	for _, t := range cqldata.GenTypes {
		if t.Name == typeName {
			return t.Kind
		}
		for _, alias := range t.Aliases {
			if alias == typeName {
				return t.Kind
			}
		}
	}
	return ""
}

// AreTypesCompatible checks if two types are compatible for comparison.
func AreTypesCompatible(type1, type2 string) bool {
	type1 = strings.ToLower(type1)
	type2 = strings.ToLower(type2)

	if type1 == type2 {
		return true
	}

	// Check direct compatibility
	if compat, ok := cqldata.GenTypeCompatibility[type1]; ok {
		for _, ct := range compat {
			if ct == type2 {
				return true
			}
		}
	}

	// Check reverse compatibility
	if compat, ok := cqldata.GenTypeCompatibility[type2]; ok {
		for _, ct := range compat {
			if ct == type1 {
				return true
			}
		}
	}

	return false
}
