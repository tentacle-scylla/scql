package cqlextract

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// FunctionDef represents a CQL function definition.
type FunctionDef struct {
	Name       string   // Function name (e.g., "now", "uuid")
	ReturnType string   // Return type (e.g., "timeuuid", "timestamp")
	Params     []string // Parameter types
	Pure       bool     // True if function is deterministic
	Aggregate  bool     // True if aggregate function
}

// ParseFunctions extracts function definitions from ScyllaDB C++ headers.
func ParseFunctions(functionsDir string) ([]FunctionDef, error) {
	var functions []FunctionDef

	files := []string{
		"time_uuid_fcts.hh",
		"uuid_fcts.hh",
		"bytes_conversion_fcts.hh",
	}

	for _, file := range files {
		path := filepath.Join(functionsDir, file)
		content, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		funcs := extractScalarFunctions(string(content))
		functions = append(functions, funcs...)
	}

	functions = append(functions, getAggregateFunctions()...)
	functions = append(functions, getSpecialFunctions()...)
	functions = deduplicateFunctions(functions)

	return functions, nil
}

// extractScalarFunctions finds make_native_scalar_function calls.
func extractScalarFunctions(content string) []FunctionDef {
	var functions []FunctionDef

	re := regexp.MustCompile(`make_native_scalar_function<(true|false)>\s*\(\s*"(\w+)"\s*,\s*(\w+)\s*,\s*\{([^}]*)\}`)

	matches := re.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) >= 5 {
			pure := match[1] == "true"
			name := match[2]
			returnType := normalizeType(match[3])
			params := parseParams(match[4])

			functions = append(functions, FunctionDef{
				Name:       name,
				ReturnType: returnType,
				Params:     params,
				Pure:       pure,
				Aggregate:  false,
			})
		}
	}

	return functions
}

// parseParams extracts parameter types from a C++ initializer list.
func parseParams(paramsStr string) []string {
	paramsStr = strings.TrimSpace(paramsStr)
	if paramsStr == "" {
		return nil
	}

	var params []string
	parts := strings.Split(paramsStr, ",")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			params = append(params, normalizeType(p))
		}
	}
	return params
}

// normalizeType converts C++ type names to CQL type names.
func normalizeType(cppType string) string {
	cppType = strings.TrimSpace(cppType)

	typeMap := map[string]string{
		"timeuuid_type":    "timeuuid",
		"uuid_type":        "uuid",
		"timestamp_type":   "timestamp",
		"long_type":        "bigint",
		"int32_type":       "int",
		"short_type":       "smallint",
		"byte_type":        "tinyint",
		"float_type":       "float",
		"double_type":      "double",
		"bytes_type":       "blob",
		"boolean_type":     "boolean",
		"utf8_type":        "text",
		"ascii_type":       "ascii",
		"inet_addr_type":   "inet",
		"simple_date_type": "date",
		"time_type":        "time",
		"duration_type":    "duration",
		"varint_type":      "varint",
		"decimal_type":     "decimal",
		"counter_type":     "counter",
		"empty_type":       "empty",
	}

	if mapped, ok := typeMap[cppType]; ok {
		return mapped
	}

	cppType = strings.TrimSuffix(cppType, "_type")
	return strings.ToLower(cppType)
}

// getAggregateFunctions returns known aggregate functions.
func getAggregateFunctions() []FunctionDef {
	return []FunctionDef{
		{Name: "count", ReturnType: "bigint", Params: []string{"any"}, Pure: true, Aggregate: true},
		{Name: "sum", ReturnType: "number", Params: []string{"number"}, Pure: true, Aggregate: true},
		{Name: "avg", ReturnType: "number", Params: []string{"number"}, Pure: true, Aggregate: true},
		{Name: "min", ReturnType: "any", Params: []string{"any"}, Pure: true, Aggregate: true},
		{Name: "max", ReturnType: "any", Params: []string{"any"}, Pure: true, Aggregate: true},
	}
}

// getSpecialFunctions returns functions that are harder to extract.
func getSpecialFunctions() []FunctionDef {
	return []FunctionDef{
		{Name: "token", ReturnType: "bigint", Params: []string{"partition_key"}, Pure: true, Aggregate: false},
		{Name: "writetime", ReturnType: "bigint", Params: []string{"column"}, Pure: true, Aggregate: false},
		{Name: "ttl", ReturnType: "int", Params: []string{"column"}, Pure: true, Aggregate: false},
		{Name: "tojson", ReturnType: "text", Params: []string{"any"}, Pure: true, Aggregate: false},
		{Name: "fromjson", ReturnType: "any", Params: []string{"text"}, Pure: true, Aggregate: false},
		{Name: "cast", ReturnType: "target_type", Params: []string{"any", "type"}, Pure: true, Aggregate: false},
		{Name: "blobasint", ReturnType: "int", Params: []string{"blob"}, Pure: true, Aggregate: false},
		{Name: "blobasbigint", ReturnType: "bigint", Params: []string{"blob"}, Pure: true, Aggregate: false},
		{Name: "blobastext", ReturnType: "text", Params: []string{"blob"}, Pure: true, Aggregate: false},
		{Name: "blobasvarchar", ReturnType: "text", Params: []string{"blob"}, Pure: true, Aggregate: false},
		{Name: "blobasascii", ReturnType: "ascii", Params: []string{"blob"}, Pure: true, Aggregate: false},
		{Name: "blobasuuid", ReturnType: "uuid", Params: []string{"blob"}, Pure: true, Aggregate: false},
		{Name: "blobasboolean", ReturnType: "boolean", Params: []string{"blob"}, Pure: true, Aggregate: false},
		{Name: "blobasdouble", ReturnType: "double", Params: []string{"blob"}, Pure: true, Aggregate: false},
		{Name: "blobasfloat", ReturnType: "float", Params: []string{"blob"}, Pure: true, Aggregate: false},
		{Name: "blobasinet", ReturnType: "inet", Params: []string{"blob"}, Pure: true, Aggregate: false},
		{Name: "intasblob", ReturnType: "blob", Params: []string{"int"}, Pure: true, Aggregate: false},
		{Name: "bigintasblob", ReturnType: "blob", Params: []string{"bigint"}, Pure: true, Aggregate: false},
		{Name: "textasblob", ReturnType: "blob", Params: []string{"text"}, Pure: true, Aggregate: false},
		{Name: "varcharasblob", ReturnType: "blob", Params: []string{"text"}, Pure: true, Aggregate: false},
		{Name: "asciiasblob", ReturnType: "blob", Params: []string{"ascii"}, Pure: true, Aggregate: false},
		{Name: "uuidasblob", ReturnType: "blob", Params: []string{"uuid"}, Pure: true, Aggregate: false},
		{Name: "booleanasblob", ReturnType: "blob", Params: []string{"boolean"}, Pure: true, Aggregate: false},
		{Name: "doubleasblob", ReturnType: "blob", Params: []string{"double"}, Pure: true, Aggregate: false},
		{Name: "floatasblob", ReturnType: "blob", Params: []string{"float"}, Pure: true, Aggregate: false},
		{Name: "inetasblob", ReturnType: "blob", Params: []string{"inet"}, Pure: true, Aggregate: false},
	}
}

// deduplicateFunctions removes duplicate functions and sorts by name.
func deduplicateFunctions(functions []FunctionDef) []FunctionDef {
	seen := make(map[string]FunctionDef)

	for _, f := range functions {
		key := f.Name
		if len(f.Params) > 0 {
			key += "(" + strings.Join(f.Params, ",") + ")"
		}

		if _, exists := seen[key]; !exists {
			seen[key] = f
		}
	}

	result := make([]FunctionDef, 0, len(seen))
	for _, f := range seen {
		result = append(result, f)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	return result
}

// GetFunctionsByReturnType groups functions by their return type.
func GetFunctionsByReturnType(functions []FunctionDef) map[string][]FunctionDef {
	result := make(map[string][]FunctionDef)
	for _, f := range functions {
		result[f.ReturnType] = append(result[f.ReturnType], f)
	}
	return result
}

// FormatFunctionSignature returns a human-readable function signature.
func FormatFunctionSignature(f FunctionDef) string {
	params := strings.Join(f.Params, ", ")
	return fmt.Sprintf("%s(%s) -> %s", f.Name, params, f.ReturnType)
}
