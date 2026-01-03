package parse

import (
	"strings"
)

// cqlKeywords contains common CQL keywords for typo detection
var cqlKeywords = []string{
	// DML
	"SELECT", "FROM", "WHERE", "AND", "OR", "IN", "INSERT", "INTO", "VALUES",
	"UPDATE", "SET", "DELETE", "USING", "TTL", "TIMESTAMP", "IF", "EXISTS",
	"NOT", "NULL", "LIMIT", "ORDER", "BY", "ASC", "DESC", "ALLOW", "FILTERING",
	"TOKEN", "CONTAINS", "KEY",

	// DDL
	"CREATE", "ALTER", "DROP", "TRUNCATE", "TABLE", "KEYSPACE", "INDEX", "TYPE",
	"MATERIALIZED", "VIEW", "FUNCTION", "AGGREGATE", "TRIGGER", "PRIMARY",
	"CLUSTERING", "COMPACT", "STORAGE", "WITH", "OPTIONS", "REPLICATION",
	"DURABLE", "WRITES", "COMMENT", "STATIC", "FROZEN", "COUNTER",

	// DCL
	"GRANT", "REVOKE", "ROLE", "USER", "PASSWORD", "SUPERUSER", "NOSUPERUSER",
	"LOGIN", "NOLOGIN", "PERMISSION", "PERMISSIONS", "ALL", "ON", "TO",

	// Types
	"ASCII", "BIGINT", "BLOB", "BOOLEAN", "DATE", "DECIMAL", "DOUBLE", "FLOAT",
	"INET", "INT", "SMALLINT", "TEXT", "TIME", "TIMEUUID", "TINYINT", "UUID",
	"VARCHAR", "VARINT", "LIST", "MAP", "SET", "TUPLE",

	// Other
	"USE", "BATCH", "BEGIN", "APPLY", "UNLOGGED", "LOGGED", "AS", "DISTINCT",
	"COUNT", "JSON", "CAST",
}

// SuggestKeyword checks if the input looks like a misspelled CQL keyword
// and returns a suggestion if a close match is found.
// Returns empty string if no suggestion.
func SuggestKeyword(input string) string {
	input = strings.ToUpper(strings.TrimSpace(input))
	if input == "" {
		return ""
	}

	// Skip very short inputs (likely not keyword typos)
	if len(input) < 4 {
		return ""
	}

	// First check for exact match (no suggestion needed)
	for _, kw := range cqlKeywords {
		if input == kw {
			return ""
		}
	}

	// Find closest keyword within threshold
	const maxDistance = 2

	bestMatch := ""
	bestDistance := maxDistance + 1

	for _, kw := range cqlKeywords {
		// Skip very short keywords (BY, OR, ON, IN, AS) - too many false positives
		if len(kw) <= 2 {
			continue
		}

		// Quick length check - if lengths differ by more than maxDistance, skip
		lenDiff := len(kw) - len(input)
		if lenDiff < 0 {
			lenDiff = -lenDiff
		}
		if lenDiff > maxDistance {
			continue
		}

		dist := levenshteinDistance(input, kw)
		if dist <= maxDistance && dist < bestDistance {
			bestDistance = dist
			bestMatch = kw
		}
	}

	if bestMatch != "" {
		return bestMatch
	}
	return ""
}

// levenshteinDistance calculates the Levenshtein distance between two strings.
// This is the minimum number of single-character edits (insertions, deletions,
// or substitutions) required to change one string into the other.
func levenshteinDistance(s1, s2 string) int {
	if s1 == s2 {
		return 0
	}
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	// Create two rows for the DP matrix
	prev := make([]int, len(s2)+1)
	curr := make([]int, len(s2)+1)

	// Initialize first row
	for j := range prev {
		prev[j] = j
	}

	// Fill the matrix
	for i := 1; i <= len(s1); i++ {
		curr[0] = i
		for j := 1; j <= len(s2); j++ {
			cost := 1
			if s1[i-1] == s2[j-1] {
				cost = 0
			}
			curr[j] = min(
				prev[j]+1,      // deletion
				curr[j-1]+1,    // insertion
				prev[j-1]+cost, // substitution
			)
		}
		prev, curr = curr, prev
	}

	return prev[len(s2)]
}

// min returns the minimum of three integers
func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}
