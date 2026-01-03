package parse

import (
	"regexp"
	"strings"

	"github.com/antlr4-go/antlr/v4"

	parser "github.com/pierre-borckmans/scql/gen/parser"
	"github.com/pierre-borckmans/scql/pkg/types"
)

// Result contains the result of parsing a CQL statement
type Result struct {
	// Input is the original CQL input string
	Input string

	// Tree is the root of the parse tree (nil if parsing failed completely)
	Tree parser.IRootContext

	// Cql is the parsed CQL statement context (nil if parsing failed)
	Cql parser.ICqlContext

	// Type is the detected statement type
	Type types.StatementType

	// Errors contains any parsing errors encountered
	Errors types.Errors

	// Tokens provides access to the token stream for formatting
	Tokens *antlr.CommonTokenStream
}

// HasErrors returns true if there were any parsing errors
func (r *Result) HasErrors() bool {
	return len(r.Errors) > 0
}

// IsValid returns true if parsing succeeded without errors
func (r *Result) IsValid() bool {
	return !r.HasErrors() && r.Cql != nil
}

// errorCollector implements antlr.ErrorListener to collect parsing errors
type errorCollector struct {
	*antlr.DefaultErrorListener
	errors types.Errors
	input  string
}

func newErrorCollector(input string) *errorCollector {
	return &errorCollector{
		DefaultErrorListener: antlr.NewDefaultErrorListener(),
		input:                input,
	}
}

func (c *errorCollector) SyntaxError(
	_ antlr.Recognizer,
	_ any,
	line, column int,
	msg string,
	_ antlr.RecognitionException,
) {
	// Transform the raw ANTLR message into a user-friendly message
	result := TransformError(msg, c.input)

	err := &types.Error{
		Line:            line,
		Column:          column,
		Message:         msg, // Keep raw message for debugging
		FriendlyMessage: result.FriendlyMessage,
		Query:           c.input,
		Suggestion:      result.Suggestion,
	}

	// If no suggestion from pattern, try Levenshtein-based keyword suggestion
	if err.Suggestion == "" {
		err.Suggestion = suggestKeywordFromError(msg, c.input)
	}

	c.errors = append(c.errors, err)
}

// suggestKeywordFromError extracts a potential typo from the error message
// and uses Levenshtein distance to suggest a correct keyword.
func suggestKeywordFromError(msg, input string) string {
	// Try to extract the mismatched token from the error message
	// Pattern: "mismatched input 'X'" or similar
	patterns := []string{
		`mismatched input '([^']+)'`,
		`extraneous input '([^']+)'`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(msg); len(matches) > 1 {
			token := matches[1]
			if suggestion := SuggestKeyword(token); suggestion != "" {
				return "Did you mean '" + suggestion + "'?"
			}
		}
	}

	// For "no viable alternative at input 'X Y Z'" - try the last word
	noViableRe := regexp.MustCompile(`no viable alternative at input '([^']+)'`)
	if matches := noViableRe.FindStringSubmatch(msg); len(matches) > 1 {
		words := strings.Fields(matches[1])
		if len(words) > 0 {
			lastWord := words[len(words)-1]
			// Clean any trailing punctuation
			lastWord = strings.TrimRight(lastWord, ";,()[]{}=<>!@#$%^&*")
			if suggestion := SuggestKeyword(lastWord); suggestion != "" {
				return "Did you mean '" + suggestion + "'?"
			}
		}
	}

	// As a fallback, scan the input for any unrecognized tokens
	// that might be typos of CQL keywords (only if they're NOT already valid keywords)
	words := strings.Fields(input)
	for _, word := range words {
		// Clean punctuation
		word = strings.TrimRight(word, ";,()[]{}=<>!@#$%^&*")
		word = strings.TrimLeft(word, "(")

		// Skip if this word is already a valid keyword
		upperWord := strings.ToUpper(word)
		isValidKeyword := false
		for _, kw := range cqlKeywords {
			if upperWord == kw {
				isValidKeyword = true
				break
			}
		}
		if isValidKeyword {
			continue
		}

		// Try to suggest a keyword for this potentially typo'd word
		if suggestion := SuggestKeyword(word); suggestion != "" {
			return "Did you mean '" + suggestion + "'?"
		}
	}

	return ""
}

// Parse parses a single CQL statement and returns a Result
func Parse(input string) *Result {
	input = strings.TrimSpace(input)

	result := &Result{
		Input: input,
	}

	// Create lexer
	inputStream := antlr.NewInputStream(input)
	lexer := parser.NewCqlLexer(inputStream)

	// Create token stream
	tokens := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	result.Tokens = tokens

	// Create parser
	p := parser.NewCqlParser(tokens)

	// Set up error collection
	collector := newErrorCollector(input)
	lexer.RemoveErrorListeners()
	lexer.AddErrorListener(collector)
	p.RemoveErrorListeners()
	p.AddErrorListener(collector)

	// Parse
	result.Tree = p.Root()
	result.Errors = collector.errors

	// Extract the CQL statement if parsing succeeded
	if result.Tree != nil {
		cqls := result.Tree.Cqls()
		if cqls != nil {
			cqlList := cqls.AllCql()
			if len(cqlList) > 0 {
				result.Cql = cqlList[0]
				result.Type = detectStatementType(result.Cql)
			}
		}
	}

	return result
}

// Multiple parses multiple CQL statements separated by semicolons
func Multiple(input string) []*Result {
	var results []*Result

	// Split by semicolons but be careful about strings
	statements := splitStatements(input)

	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}

		// Ensure statement ends with semicolon for parsing
		if !strings.HasSuffix(stmt, ";") {
			stmt += ";"
		}

		results = append(results, Parse(stmt))
	}

	return results
}

// splitStatements splits CQL input into individual statements
// It handles quoted strings properly to avoid splitting on semicolons inside strings
func splitStatements(input string) []string {
	var statements []string
	var current strings.Builder
	inString := false
	stringChar := rune(0)

	for i, ch := range input {
		if !inString {
			if ch == '\'' || ch == '"' {
				inString = true
				stringChar = ch
				current.WriteRune(ch)
			} else if ch == ';' {
				stmt := strings.TrimSpace(current.String())
				if stmt != "" {
					statements = append(statements, stmt+";")
				}
				current.Reset()
			} else if ch == '-' && i+1 < len(input) && input[i+1] == '-' {
				// Skip line comments
				for i < len(input) && input[i] != '\n' {
					i++
				}
			} else {
				current.WriteRune(ch)
			}
		} else {
			current.WriteRune(ch)
			if ch == stringChar {
				// Check for escaped quote
				if i+1 < len(input) && rune(input[i+1]) == stringChar {
					continue
				}
				inString = false
				stringChar = 0
			}
		}
	}

	// Handle any remaining content
	stmt := strings.TrimSpace(current.String())
	if stmt != "" {
		statements = append(statements, stmt)
	}

	return statements
}

// IsValid returns true if the CQL input is syntactically valid
func IsValid(input string) bool {
	return !Parse(input).HasErrors()
}

// detectStatementType determines the statement type from the parse tree
func detectStatementType(ctx parser.ICqlContext) types.StatementType {
	if ctx == nil {
		return types.StatementUnknown
	}

	if ctx.Select_() != nil {
		return types.StatementSelect
	}
	if ctx.Insert() != nil {
		return types.StatementInsert
	}
	if ctx.Update() != nil {
		return types.StatementUpdate
	}
	if ctx.Delete_() != nil {
		return types.StatementDelete
	}
	if ctx.ApplyBatch() != nil {
		return types.StatementBatch
	}
	if ctx.CreateKeyspace() != nil {
		return types.StatementCreateKeyspace
	}
	if ctx.AlterKeyspace() != nil {
		return types.StatementAlterKeyspace
	}
	if ctx.DropKeyspace() != nil {
		return types.StatementDropKeyspace
	}
	if ctx.CreateTable() != nil {
		return types.StatementCreateTable
	}
	if ctx.AlterTable() != nil {
		return types.StatementAlterTable
	}
	if ctx.DropTable() != nil {
		return types.StatementDropTable
	}
	if ctx.Truncate() != nil {
		return types.StatementTruncate
	}
	if ctx.CreateIndex() != nil {
		return types.StatementCreateIndex
	}
	if ctx.DropIndex() != nil {
		return types.StatementDropIndex
	}
	if ctx.CreateMaterializedView() != nil {
		return types.StatementCreateMaterializedView
	}
	if ctx.AlterMaterializedView() != nil {
		return types.StatementAlterMaterializedView
	}
	if ctx.DropMaterializedView() != nil {
		return types.StatementDropMaterializedView
	}
	if ctx.CreateType() != nil {
		return types.StatementCreateType
	}
	if ctx.AlterType() != nil {
		return types.StatementAlterType
	}
	if ctx.DropType() != nil {
		return types.StatementDropType
	}
	if ctx.CreateFunction() != nil {
		return types.StatementCreateFunction
	}
	if ctx.DropFunction() != nil {
		return types.StatementDropFunction
	}
	if ctx.CreateAggregate() != nil {
		return types.StatementCreateAggregate
	}
	if ctx.DropAggregate() != nil {
		return types.StatementDropAggregate
	}
	if ctx.CreateTrigger() != nil {
		return types.StatementCreateTrigger
	}
	if ctx.DropTrigger() != nil {
		return types.StatementDropTrigger
	}
	if ctx.CreateRole() != nil {
		return types.StatementCreateRole
	}
	if ctx.AlterRole() != nil {
		return types.StatementAlterRole
	}
	if ctx.DropRole() != nil {
		return types.StatementDropRole
	}
	if ctx.CreateUser() != nil {
		return types.StatementCreateUser
	}
	if ctx.AlterUser() != nil {
		return types.StatementAlterUser
	}
	if ctx.DropUser() != nil {
		return types.StatementDropUser
	}
	if ctx.Grant() != nil {
		return types.StatementGrant
	}
	if ctx.Revoke() != nil {
		return types.StatementRevoke
	}
	if ctx.ListRoles() != nil {
		return types.StatementListRoles
	}
	if ctx.ListPermissions() != nil {
		return types.StatementListPermissions
	}
	if ctx.Use_() != nil {
		return types.StatementUse
	}
	if ctx.PruneMaterializedView() != nil {
		return types.StatementPruneMaterializedView
	}

	return types.StatementUnknown
}
