package analyze

import (
	"strconv"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	parser "github.com/pierre-borckmans/scql/gen/parser"
	"github.com/pierre-borckmans/scql/pkg/parse"
	"github.com/pierre-borckmans/scql/pkg/types"
)

// ExtractReferences parses a CQL query and extracts all schema references.
func ExtractReferences(cql string) (*References, types.StatementType, types.Errors) {
	result := parse.Parse(cql)
	if result.Tree == nil {
		return NewReferences(), types.StatementUnknown, result.Errors
	}

	extractor := newReferenceExtractor()
	antlr.ParseTreeWalkerDefault.Walk(extractor, result.Tree)

	return extractor.refs, result.Type, result.Errors
}

// referenceExtractor implements the ANTLR listener to extract schema references.
type referenceExtractor struct {
	*parser.BaseCqlParserListener
	refs *References

	// Track context for column references
	inSelect  bool
	inWhere   bool
	inOrderBy bool
	inUpdate  bool
	inInsert  bool
	inGroupBy bool
}

func newReferenceExtractor() *referenceExtractor {
	return &referenceExtractor{
		BaseCqlParserListener: &parser.BaseCqlParserListener{},
		refs:                  NewReferences(),
	}
}

// Table reference extraction

func (e *referenceExtractor) EnterTable(ctx *parser.TableContext) {
	// Table can be keyspace.table or just table (used by INSERT, UPDATE, CREATE, etc.)
	e.extractTableRef(ctx.GetText())
}

func (e *referenceExtractor) EnterFromSpecElement(ctx *parser.FromSpecElementContext) {
	// FromSpecElement is used by SELECT and DELETE for table references
	// It contains OBJECT_NAME nodes with optional DOT separator
	e.extractTableRef(ctx.GetText())
}

func (e *referenceExtractor) extractTableRef(text string) {
	if text == "" {
		return
	}

	// Remove quotes
	text = strings.Trim(text, "\"")

	parts := strings.Split(text, ".")
	if len(parts) == 2 {
		e.refs.Keyspace = strings.Trim(parts[0], "\"")
		e.refs.Table = strings.Trim(parts[1], "\"")
	} else {
		e.refs.Table = strings.Trim(text, "\"")
	}
}

// SELECT statement handling

func (e *referenceExtractor) EnterSelect_(ctx *parser.Select_Context) {
	e.inSelect = true
}

func (e *referenceExtractor) ExitSelect_(ctx *parser.Select_Context) {
	e.inSelect = false
}

func (e *referenceExtractor) EnterSelectElements(ctx *parser.SelectElementsContext) {
	// Check for SELECT *
	if ctx.GetText() == "*" {
		e.refs.SelectColumns = append(e.refs.SelectColumns, "*")
	}
}

func (e *referenceExtractor) EnterSelectElement(ctx *parser.SelectElementContext) {
	// Extract column name from select element
	// This can be complex (function calls, aliases, etc.)
	// For now, extract simple column references
}

// WHERE clause handling

func (e *referenceExtractor) EnterWhereSpec(ctx *parser.WhereSpecContext) {
	e.inWhere = true
}

func (e *referenceExtractor) ExitWhereSpec(ctx *parser.WhereSpecContext) {
	e.inWhere = false
}

// ORDER BY handling

func (e *referenceExtractor) EnterOrderSpec(ctx *parser.OrderSpecContext) {
	e.inOrderBy = true
}

func (e *referenceExtractor) ExitOrderSpec(ctx *parser.OrderSpecContext) {
	e.inOrderBy = false
}

func (e *referenceExtractor) EnterOrderSpecElement(ctx *parser.OrderSpecElementContext) {
	// Extract the column name from ORDER BY
	if objName := ctx.OBJECT_NAME(); objName != nil {
		colName := extractColumnName(objName.GetText())
		e.refs.OrderByColumns = append(e.refs.OrderByColumns, colName)
		e.addColumn(colName)
	}
}

// UPDATE statement handling

func (e *referenceExtractor) EnterUpdate(ctx *parser.UpdateContext) {
	e.inUpdate = true
}

func (e *referenceExtractor) ExitUpdate(ctx *parser.UpdateContext) {
	e.inUpdate = false
}

func (e *referenceExtractor) EnterAssignmentElement(ctx *parser.AssignmentElementContext) {
	// Extract column being assigned (first ColumnRef is the target column)
	if colRefs := ctx.AllColumnRef(); len(colRefs) > 0 {
		colName := extractColumnName(colRefs[0].GetText())
		e.refs.UpdateColumns = append(e.refs.UpdateColumns, colName)
		e.addColumn(colName)
	}
}

// INSERT statement handling

func (e *referenceExtractor) EnterInsert(ctx *parser.InsertContext) {
	e.inInsert = true
}

func (e *referenceExtractor) ExitInsert(ctx *parser.InsertContext) {
	e.inInsert = false
}

func (e *referenceExtractor) EnterInsertColumnSpec(ctx *parser.InsertColumnSpecContext) {
	// Extract columns from INSERT column list
	if colList := ctx.ColumnList(); colList != nil {
		for _, col := range colList.AllColumn() {
			colName := extractColumnName(col.GetText())
			e.refs.InsertColumns = append(e.refs.InsertColumns, colName)
			e.addColumn(colName)
		}
	}
}

// Column reference handling

func (e *referenceExtractor) EnterColumn(ctx *parser.ColumnContext) {
	colName := extractColumnName(ctx.GetText())
	if colName == "" {
		return
	}

	e.addColumn(colName)

	// Categorize based on context
	if e.inWhere {
		e.refs.WhereColumns = append(e.refs.WhereColumns, colName)
	} else if e.inSelect && !e.inWhere && !e.inOrderBy {
		// Only add to SelectColumns if we're in SELECT but not in a subclause
		if !contains(e.refs.SelectColumns, colName) && !contains(e.refs.SelectColumns, "*") {
			e.refs.SelectColumns = append(e.refs.SelectColumns, colName)
		}
	}
}

func (e *referenceExtractor) EnterColumnRef(ctx *parser.ColumnRefContext) {
	// ColumnRef can be keyspace.table.column, table.column, or column
	text := ctx.GetText()
	parts := strings.Split(text, ".")
	var colName string
	switch len(parts) {
	case 3:
		// keyspace.table.column
		colName = parts[2]
	case 2:
		// table.column
		colName = parts[1]
	default:
		colName = text
	}

	colName = extractColumnName(colName)
	if colName == "" {
		return
	}

	e.addColumn(colName)

	if e.inWhere {
		e.refs.WhereColumns = append(e.refs.WhereColumns, colName)
	} else if e.inSelect {
		if !contains(e.refs.SelectColumns, colName) {
			e.refs.SelectColumns = append(e.refs.SelectColumns, colName)
		}
	}
}

// Function call handling

func (e *referenceExtractor) EnterFunctionCall(ctx *parser.FunctionCallContext) {
	// Extract function name
	var fnName string
	if objName := ctx.OBJECT_NAME(); objName != nil {
		fnName = strings.ToLower(objName.GetText())
	} else if uuid := ctx.K_UUID(); uuid != nil {
		fnName = "uuid"
	} else if token := ctx.KwToken(); token != nil {
		fnName = "token"
	} else if ttl := ctx.KwTtl(); ttl != nil {
		fnName = "ttl"
	} else if writetime := ctx.KwWritetime(); writetime != nil {
		fnName = "writetime"
	}

	if fnName == "" {
		return
	}

	// Add to simple function list (backward compat)
	if !contains(e.refs.Functions, fnName) {
		e.refs.Functions = append(e.refs.Functions, fnName)
	}

	// Create detailed function call info
	fc := &FunctionCall{
		Name: fnName,
		Position: &Position{
			Line:   ctx.GetStart().GetLine(),
			Column: ctx.GetStart().GetColumn(),
			Offset: ctx.GetStart().GetStart(),
		},
	}

	// Check for star argument (e.g., count(*))
	if ctx.STAR() != nil {
		fc.HasStar = true
		fc.ArgCount = 1
	} else if args := ctx.FunctionArgs(); args != nil {
		// Count arguments by counting elements in FunctionArgs
		// Arguments can be constants, column refs, or nested function calls
		argsCtx := args.(*parser.FunctionArgsContext)
		argCount := 0
		argCount += len(argsCtx.AllConstant())
		argCount += len(argsCtx.AllColumnRef())
		argCount += len(argsCtx.AllFunctionCall())
		argCount += len(argsCtx.AllQualifiedFunctionCall())
		fc.ArgCount = argCount
	}

	e.refs.FunctionCalls = append(e.refs.FunctionCalls, fc)
}

// ALLOW FILTERING handling

func (e *referenceExtractor) EnterAllowFilteringSpec(ctx *parser.AllowFilteringSpecContext) {
	e.refs.HasAllowFiltering = true
}

// LIMIT handling

func (e *referenceExtractor) EnterLimitSpec(ctx *parser.LimitSpecContext) {
	// Extract limit value
	text := ctx.GetText()
	// Remove "LIMIT" prefix
	text = strings.TrimPrefix(strings.ToUpper(text), "LIMIT")
	text = strings.TrimSpace(text)
	if val, err := strconv.Atoi(text); err == nil {
		e.refs.Limit = val
	}
}

// Helper methods

func (e *referenceExtractor) addColumn(name string) {
	name = extractColumnName(name)
	if name != "" && !contains(e.refs.Columns, name) {
		e.refs.Columns = append(e.refs.Columns, name)
	}
}

func extractColumnName(text string) string {
	// Remove quotes and whitespace
	text = strings.TrimSpace(text)
	text = strings.Trim(text, "\"")
	// Handle array access like column[0]
	if idx := strings.Index(text, "["); idx != -1 {
		text = text[:idx]
	}
	return text
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
