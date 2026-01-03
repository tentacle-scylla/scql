package cqlextract

import (
	"os"
	"regexp"
	"sort"
	"strings"
)

// TypeDef represents a CQL type definition.
type TypeDef struct {
	Name    string   // Primary type name (e.g., "text")
	Aliases []string // Alternative names (e.g., "varchar" for text)
	Kind    string   // Type category (e.g., "string", "integer", "uuid")
}

// TypeData holds extracted type information.
type TypeData struct {
	Types         []TypeDef           // All types
	Compatibility map[string][]string // type -> compatible types
}

// ParseTypes extracts type definitions from ScyllaDB source.
func ParseTypes(typesPath string) (*TypeData, error) {
	content, err := os.ReadFile(typesPath)
	if err != nil {
		return getDefaultTypeData(), nil
	}

	data := &TypeData{
		Compatibility: make(map[string][]string),
	}

	data.Types = extractTypes(string(content))

	if len(data.Types) == 0 {
		data.Types = getDefaultTypes()
	}

	data.Compatibility = extractCompatibility(string(content))
	if len(data.Compatibility) == 0 {
		data.Compatibility = getDefaultCompatibility()
	}

	return data, nil
}

// extractTypes attempts to extract type names from types.cc.
func extractTypes(content string) []TypeDef {
	var types []TypeDef

	typePattern := regexp.MustCompile(`(\w+)_type(?:\s|,|;|\))`)

	matches := typePattern.FindAllStringSubmatch(content, -1)
	seen := make(map[string]bool)

	for _, match := range matches {
		if len(match) > 1 {
			typeName := match[1]
			if strings.HasSuffix(typeName, "_impl") ||
				strings.HasPrefix(typeName, "abstract_") ||
				typeName == "data" ||
				typeName == "native" {
				continue
			}

			if !seen[typeName] {
				seen[typeName] = true
				types = append(types, TypeDef{
					Name: normalizeTypeName(typeName),
					Kind: categorizeType(typeName),
				})
			}
		}
	}

	return types
}

// normalizeTypeName converts C++ type names to CQL names.
func normalizeTypeName(name string) string {
	switch name {
	case "utf8":
		return "text"
	case "long":
		return "bigint"
	case "int32":
		return "int"
	case "short":
		return "smallint"
	case "byte":
		return "tinyint"
	case "bytes":
		return "blob"
	case "inet_addr":
		return "inet"
	case "simple_date":
		return "date"
	}
	return name
}

// categorizeType determines the type category.
func categorizeType(name string) string {
	switch name {
	case "int32", "long", "short", "byte", "varint", "decimal", "float", "double", "counter":
		return "numeric"
	case "utf8", "ascii", "varchar":
		return "string"
	case "uuid", "timeuuid":
		return "uuid"
	case "timestamp", "date", "time", "duration", "simple_date":
		return "temporal"
	case "boolean":
		return "boolean"
	case "bytes", "blob":
		return "binary"
	case "inet_addr", "inet":
		return "network"
	case "list", "set", "map", "tuple", "vector":
		return "collection"
	default:
		return "other"
	}
}

// extractCompatibility extracts type compatibility rules from source.
func extractCompatibility(content string) map[string][]string {
	return getDefaultCompatibility()
}

// getDefaultTypeData returns a complete default TypeData.
func getDefaultTypeData() *TypeData {
	return &TypeData{
		Types:         getDefaultTypes(),
		Compatibility: getDefaultCompatibility(),
	}
}

// getDefaultTypes returns the known CQL types.
func getDefaultTypes() []TypeDef {
	types := []TypeDef{
		{Name: "int", Aliases: nil, Kind: "numeric"},
		{Name: "bigint", Aliases: []string{"long"}, Kind: "numeric"},
		{Name: "smallint", Aliases: nil, Kind: "numeric"},
		{Name: "tinyint", Aliases: nil, Kind: "numeric"},
		{Name: "varint", Aliases: nil, Kind: "numeric"},
		{Name: "float", Aliases: nil, Kind: "numeric"},
		{Name: "double", Aliases: nil, Kind: "numeric"},
		{Name: "decimal", Aliases: nil, Kind: "numeric"},
		{Name: "counter", Aliases: nil, Kind: "numeric"},

		{Name: "text", Aliases: []string{"varchar"}, Kind: "string"},
		{Name: "ascii", Aliases: nil, Kind: "string"},

		{Name: "uuid", Aliases: nil, Kind: "uuid"},
		{Name: "timeuuid", Aliases: nil, Kind: "uuid"},

		{Name: "timestamp", Aliases: nil, Kind: "temporal"},
		{Name: "date", Aliases: nil, Kind: "temporal"},
		{Name: "time", Aliases: nil, Kind: "temporal"},
		{Name: "duration", Aliases: nil, Kind: "temporal"},

		{Name: "boolean", Aliases: nil, Kind: "boolean"},
		{Name: "blob", Aliases: []string{"bytes"}, Kind: "binary"},
		{Name: "inet", Aliases: nil, Kind: "network"},

		{Name: "list", Aliases: nil, Kind: "collection"},
		{Name: "set", Aliases: nil, Kind: "collection"},
		{Name: "map", Aliases: nil, Kind: "collection"},
		{Name: "tuple", Aliases: nil, Kind: "collection"},
		{Name: "frozen", Aliases: nil, Kind: "collection"},
		{Name: "vector", Aliases: nil, Kind: "collection"},

		{Name: "empty", Aliases: nil, Kind: "special"},
	}

	sort.Slice(types, func(i, j int) bool {
		return types[i].Name < types[j].Name
	})

	return types
}

// getDefaultCompatibility returns known type compatibility rules.
func getDefaultCompatibility() map[string][]string {
	return map[string][]string{
		"text":    {"ascii", "blob"},
		"ascii":   {"text", "blob"},
		"varchar": {"text", "ascii", "blob"},

		"uuid":     {"timeuuid"},
		"timeuuid": {"uuid"},

		"timestamp": {"date", "bigint"},
		"date":      {"timestamp"},

		"bigint":   {"int", "smallint", "tinyint", "varint"},
		"int":      {"smallint", "tinyint"},
		"smallint": {"tinyint"},
		"varint":   {"bigint", "int", "smallint", "tinyint"},
		"double":   {"float"},
		"decimal":  {"float", "double", "bigint", "int", "smallint", "tinyint", "varint"},

		"blob": {"text", "ascii", "int", "bigint", "uuid", "timeuuid", "timestamp"},
	}
}

// GetTypesByKind groups types by their kind.
func GetTypesByKind(types []TypeDef) map[string][]TypeDef {
	result := make(map[string][]TypeDef)
	for _, t := range types {
		result[t.Kind] = append(result[t.Kind], t)
	}
	return result
}

// IsCompatible checks if two types are compatible.
func IsCompatible(compatibility map[string][]string, type1, type2 string) bool {
	if type1 == type2 {
		return true
	}

	if compatTypes, ok := compatibility[type1]; ok {
		for _, ct := range compatTypes {
			if ct == type2 {
				return true
			}
		}
	}

	if compatTypes, ok := compatibility[type2]; ok {
		for _, ct := range compatTypes {
			if ct == type1 {
				return true
			}
		}
	}

	return false
}
