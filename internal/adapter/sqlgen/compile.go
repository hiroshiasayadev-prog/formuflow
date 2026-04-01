package sqlgen

import (
	"fmt"
	"strings"

	"github.com/hiroshiasayadev-prog/formuflow/internal/domain/ir/expr"
	"github.com/hiroshiasayadev-prog/formuflow/internal/domain/ir/rel"
)

// CompileResult holds the output of a successful compilation.
// SQL is the final WITH ... SELECT statement ready for DuckDB execution.
// Diags contains warnings or errors encountered during compilation;
// a non-empty Diags does not necessarily mean SQL is empty.
type CompileResult struct {
	SQL   string
	Diags []Diagnostic
}

// Diagnostic represents a single compiler message tied to an App layer node.
// DiagAnchor is the ID of the App node that caused the diagnostic,
// allowing the frontend to highlight the offending node.
type Diagnostic struct {
	DiagAnchor string
	Message    string
	Severity   string // "error" | "warning"
}

// compiledRel is the compiler's internal representation of a compiled RelNode.
// It carries the CTE name assigned to this node, the SQL body of that CTE,
// the output schema (used for column resolution in parent nodes),
// and the DiagAnchor for error reporting.
type compiledRel struct {
	CTEName    string
	SQL        string
	Schema     rel.Schema
	DiagAnchor string
}

// compileContext holds mutable state threaded through the recursive compilation.
type compileContext struct {
	provider rel.SchemaProvider
	cteSeq   int               // counter for generating CTE names c1, c2, ...
	ctes     []string          // accumulated CTE definitions in order
	diagMap  map[string]string // CTEName → DiagAnchor
	diags    []Diagnostic
}

func newCompileContext(provider rel.SchemaProvider) *compileContext {
	return &compileContext{
		provider: provider,
		diagMap:  make(map[string]string),
	}
}

// nextCTEName generates the next sequential CTE name (c1, c2, ...).
func (ctx *compileContext) nextCTEName() string {
	ctx.cteSeq++
	return fmt.Sprintf("c%d", ctx.cteSeq)
}

func (ctx *compileContext) addDiag(anchor, message, severity string) {
	ctx.diags = append(ctx.diags, Diagnostic{
		DiagAnchor: anchor,
		Message:    message,
		Severity:   severity,
	})
}

// Compile is the entry point. It takes a RelNode tree and a SchemaProvider
// and returns a CompileResult containing the full SQL and any diagnostics.
func Compile(node rel.RelNode, provider rel.SchemaProvider) (CompileResult, error) {
	ctx := newCompileContext(provider)

	compiled, err := compileRel(ctx, node)
	if err != nil {
		return CompileResult{}, err
	}

	// Assemble: WITH c1 AS (...), c2 AS (...) SELECT * FROM cN
	var sb strings.Builder
	sb.WriteString("WITH\n")
	for i, cte := range ctx.ctes {
		if i > 0 {
			sb.WriteString(",\n")
		}
		sb.WriteString(cte)
	}
	sb.WriteString("\nSELECT * FROM ")
	sb.WriteString(compiled.CTEName)

	return CompileResult{
		SQL:   sb.String(),
		Diags: ctx.diags,
	}, nil
}

// compileRel recursively compiles a RelNode into a compiledRel.
// CTEs are accumulated in ctx.ctes in evaluation order (child before parent).
func compileRel(ctx *compileContext, node rel.RelNode) (compiledRel, error) {
	switch n := node.(type) {
	case rel.TableScan:
		return compileTableScan(ctx, n)
	case rel.Filter:
		return compileFilter(ctx, n)
	case rel.DeriveColumn:
		return compileDeriveColumn(ctx, n)
	case rel.Project:
		return compileProject(ctx, n)
	case rel.Join:
		return compileJoin(ctx, n)
	case rel.Aggregate:
		return compileAggregate(ctx, n)
	default:
		return compiledRel{}, fmt.Errorf("unsupported RelNode type: %T", node)
	}
}

// compileTableScan compiles a TableScan node.
//
// Generates:
//
//	cN AS (SELECT * FROM <table>)
func compileTableScan(ctx *compileContext, node rel.TableScan) (compiledRel, error) {
	if node.Table.Name == "" {
		ctx.addDiag(node.DiagAnchor, "table name must not be empty", "error")
		return compiledRel{}, fmt.Errorf("TableScan.Table.Name is empty (anchor: %s)", node.DiagAnchor)
	}

	schema, err := ctx.provider.GetSchema(node.Table.Name)
	if err != nil {
		ctx.addDiag(node.DiagAnchor, fmt.Sprintf("failed to resolve schema for table %q: %v", node.Table.Name, err), "error")
		return compiledRel{}, fmt.Errorf("schema resolution failed for %q: %w", node.Table.Name, err)
	}

	cteName := ctx.nextCTEName()
	sql := fmt.Sprintf("%s AS (\n  SELECT * FROM %s\n)", cteName, node.Table.Name)
	ctx.ctes = append(ctx.ctes, sql)
	ctx.diagMap[cteName] = node.DiagAnchor

	return compiledRel{
		CTEName:    cteName,
		SQL:        sql,
		Schema:     schema,
		DiagAnchor: node.DiagAnchor,
	}, nil
}

// compileFilter compiles a Filter node.
//
// Generates:
//
//	cN AS (SELECT * FROM <input_cte> WHERE <predicate>)
func compileFilter(ctx *compileContext, node rel.Filter) (compiledRel, error) {
	input, err := compileRel(ctx, node.Input)
	if err != nil {
		return compiledRel{}, err
	}

	predSQL, err := compileExpr(node.Predicate, input.Schema)
	if err != nil {
		ctx.addDiag(node.DiagAnchor, fmt.Sprintf("invalid predicate: %v", err), "error")
		return compiledRel{}, fmt.Errorf("Filter predicate compilation failed (anchor: %s): %w", node.DiagAnchor, err)
	}

	cteName := ctx.nextCTEName()
	sql := fmt.Sprintf("%s AS (\n  SELECT * FROM %s\n  WHERE %s\n)", cteName, input.CTEName, predSQL)
	ctx.ctes = append(ctx.ctes, sql)
	ctx.diagMap[cteName] = node.DiagAnchor

	// Filter does not change the schema.
	return compiledRel{
		CTEName:    cteName,
		SQL:        sql,
		Schema:     input.Schema,
		DiagAnchor: node.DiagAnchor,
	}, nil
}

// compileProject compiles a Project node.
//
// Generates:
//
//	cN AS (SELECT <col1>, <col2>, ... FROM <input_cte>)
//
// Guards:
//   - Columns must not be empty.
//   - Every listed column must exist in the input schema (case-sensitive exact match).
func compileProject(ctx *compileContext, node rel.Project) (compiledRel, error) {
	if len(node.Columns) == 0 {
		ctx.addDiag(node.DiagAnchor, "project column list must not be empty", "error")
		return compiledRel{}, fmt.Errorf("Project.Columns is empty (anchor: %s)", node.DiagAnchor)
	}

	input, err := compileRel(ctx, node.Input)
	if err != nil {
		return compiledRel{}, err
	}

	// Validate all columns exist in the input schema before generating any SQL.
	for _, name := range node.Columns {
		if err := resolveColumn(name, input.Schema); err != nil {
			msg := fmt.Sprintf("column %q not found in input schema", name)
			ctx.addDiag(node.DiagAnchor, msg, "error")
			return compiledRel{}, fmt.Errorf("Project column resolution failed (anchor: %s): %s", node.DiagAnchor, msg)
		}
	}

	cteName := ctx.nextCTEName()
	sql := fmt.Sprintf("%s AS (\n  SELECT %s\n  FROM %s\n)", cteName, strings.Join(node.Columns, ", "), input.CTEName)
	ctx.ctes = append(ctx.ctes, sql)
	ctx.diagMap[cteName] = node.DiagAnchor

	// Output schema contains only the selected columns, in the order specified.
	outputSchema := make(rel.Schema, len(node.Columns))
	for i, name := range node.Columns {
		for _, col := range input.Schema {
			if col.Name == name {
				outputSchema[i] = col
				break
			}
		}
	}

	return compiledRel{
		CTEName:    cteName,
		SQL:        sql,
		Schema:     outputSchema,
		DiagAnchor: node.DiagAnchor,
	}, nil
}

// compileJoin compiles a Join node.
//
// CROSS generates:
//
//	cN AS (SELECT * FROM <left_cte>, <right_cte>)
//
// INNER/LEFT generates:
//
//	cN AS (SELECT * FROM <left_cte> INNER JOIN <right_cte> ON <predicate>)
//
// Guards:
//   - Left and Right column names must not overlap.
//   - On must be nil for CROSS, non-nil for INNER/LEFT.
func compileJoin(ctx *compileContext, node rel.Join) (compiledRel, error) {
	left, err := compileRel(ctx, node.Left)
	if err != nil {
		return compiledRel{}, err
	}
	right, err := compileRel(ctx, node.Right)
	if err != nil {
		return compiledRel{}, err
	}

	// Guard: column name conflicts between Left and Right.
	for _, lc := range left.Schema {
		for _, rc := range right.Schema {
			if lc.Name == rc.Name {
				msg := fmt.Sprintf("column %q exists in both left and right inputs", lc.Name)
				ctx.addDiag(node.DiagAnchor, msg, "error")
				return compiledRel{}, fmt.Errorf("Join column conflict (anchor: %s): %s", node.DiagAnchor, msg)
			}
		}
	}

	cteName := ctx.nextCTEName()

	var sql string
	switch node.Kind {
	case rel.JoinCross:
		if node.On != nil {
			ctx.addDiag(node.DiagAnchor, "CROSS JOIN must not have an ON predicate", "error")
			return compiledRel{}, fmt.Errorf("Join CROSS with On set (anchor: %s)", node.DiagAnchor)
		}
		sql = fmt.Sprintf("%s AS (\n  SELECT * FROM %s, %s\n)", cteName, left.CTEName, right.CTEName)
	case rel.JoinInner, rel.JoinLeft:
		if node.On == nil {
			ctx.addDiag(node.DiagAnchor, "INNER/LEFT JOIN requires an ON predicate", "error")
			return compiledRel{}, fmt.Errorf("Join %s missing On (anchor: %s)", node.Kind, node.DiagAnchor)
		}
		onSQL, err := compileExpr(node.On, append(left.Schema, right.Schema...))
		if err != nil {
			ctx.addDiag(node.DiagAnchor, fmt.Sprintf("invalid ON predicate: %v", err), "error")
			return compiledRel{}, fmt.Errorf("Join ON compilation failed (anchor: %s): %w", node.DiagAnchor, err)
		}
		keyword := "INNER JOIN"
		if node.Kind == rel.JoinLeft {
			keyword = "LEFT JOIN"
		}
		sql = fmt.Sprintf("%s AS (\n  SELECT * FROM %s\n  %s %s ON %s\n)", cteName, left.CTEName, keyword, right.CTEName, onSQL)
	default:
		return compiledRel{}, fmt.Errorf("unsupported JoinKind: %q (anchor: %s)", node.Kind, node.DiagAnchor)
	}

	ctx.ctes = append(ctx.ctes, sql)
	ctx.diagMap[cteName] = node.DiagAnchor

	outputSchema := append(append(rel.Schema{}, left.Schema...), right.Schema...)
	return compiledRel{
		CTEName:    cteName,
		SQL:        sql,
		Schema:     outputSchema,
		DiagAnchor: node.DiagAnchor,
	}, nil
}

// compileAggregate compiles an Aggregate node.
//
// Generates:
//
//	cN AS (SELECT <groupby_cols>, AGG(col) AS alias, ... FROM <input_cte> GROUP BY <groupby_cols>)
//
// If GroupBy is empty, GROUP BY clause is omitted (full-table aggregation).
//
// Guards:
//   - Aggs must not be empty.
//   - Each AggExpr.Col must exist in the input schema.
//   - Each GroupBy column must exist in the input schema.
func compileAggregate(ctx *compileContext, node rel.Aggregate) (compiledRel, error) {
	if len(node.Aggs) == 0 {
		ctx.addDiag(node.DiagAnchor, "aggregate must have at least one aggregation", "error")
		return compiledRel{}, fmt.Errorf("Aggregate.Aggs is empty (anchor: %s)", node.DiagAnchor)
	}

	input, err := compileRel(ctx, node.Input)
	if err != nil {
		return compiledRel{}, err
	}

	// Guard: GroupBy columns must exist in input schema.
	for _, col := range node.GroupBy {
		if err := resolveColumn(col, input.Schema); err != nil {
			msg := fmt.Sprintf("GROUP BY column %q not found in input schema", col)
			ctx.addDiag(node.DiagAnchor, msg, "error")
			return compiledRel{}, fmt.Errorf("Aggregate GroupBy resolution failed (anchor: %s): %s", node.DiagAnchor, msg)
		}
	}

	// Guard: Agg cols must exist in input schema.
	for _, agg := range node.Aggs {
		if err := resolveColumn(agg.Col, input.Schema); err != nil {
			msg := fmt.Sprintf("aggregate column %q not found in input schema", agg.Col)
			ctx.addDiag(node.DiagAnchor, msg, "error")
			return compiledRel{}, fmt.Errorf("Aggregate col resolution failed (anchor: %s): %s", node.DiagAnchor, msg)
		}
	}

	// Build SELECT clause: groupby cols + agg exprs.
	selectCols := make([]string, 0, len(node.GroupBy)+len(node.Aggs))
	for _, col := range node.GroupBy {
		selectCols = append(selectCols, col)
	}
	for _, agg := range node.Aggs {
		selectCols = append(selectCols, fmt.Sprintf("%s(%s) AS %s", agg.Func, agg.Col, agg.Alias))
	}

	cteName := ctx.nextCTEName()

	var sql string
	if len(node.GroupBy) == 0 {
		sql = fmt.Sprintf("%s AS (\n  SELECT %s\n  FROM %s\n)", cteName, strings.Join(selectCols, ", "), input.CTEName)
	} else {
		sql = fmt.Sprintf("%s AS (\n  SELECT %s\n  FROM %s\n  GROUP BY %s\n)", cteName, strings.Join(selectCols, ", "), input.CTEName, strings.Join(node.GroupBy, ", "))
	}

	ctx.ctes = append(ctx.ctes, sql)
	ctx.diagMap[cteName] = node.DiagAnchor

	// Output schema: GroupBy cols + Aggs alias cols.
	outputSchema := make(rel.Schema, 0, len(node.GroupBy)+len(node.Aggs))
	for _, col := range node.GroupBy {
		for _, c := range input.Schema {
			if c.Name == col {
				outputSchema = append(outputSchema, c)
				break
			}
		}
	}
	for _, agg := range node.Aggs {
		outputSchema = append(outputSchema, rel.Column{Name: agg.Alias, Type: ""})
	}

	return compiledRel{
		CTEName:    cteName,
		SQL:        sql,
		Schema:     outputSchema,
		DiagAnchor: node.DiagAnchor,
	}, nil
}

// compileDeriveColumn compiles a DeriveColumn node.
//
// Generates:
//
//	cN AS (SELECT *, <expr> AS <column_name> FROM <input_cte>)
//
// Guards:
//   - ColumnName must not be empty.
//   - ColumnName must not duplicate any column already in the input schema.
func compileDeriveColumn(ctx *compileContext, node rel.DeriveColumn) (compiledRel, error) {
	if node.ColumnName == "" {
		ctx.addDiag(node.DiagAnchor, "derived column name must not be empty", "error")
		return compiledRel{}, fmt.Errorf("DeriveColumn.ColumnName is empty (anchor: %s)", node.DiagAnchor)
	}

	input, err := compileRel(ctx, node.Input)
	if err != nil {
		return compiledRel{}, err
	}

	// Guard: ColumnName must not duplicate an existing column.
	if err := resolveColumn(node.ColumnName, input.Schema); err == nil {
		msg := fmt.Sprintf("column %q already exists in input schema; overwriting is not allowed", node.ColumnName)
		ctx.addDiag(node.DiagAnchor, msg, "error")
		return compiledRel{}, fmt.Errorf("DeriveColumn conflict (anchor: %s): %s", node.DiagAnchor, msg)
	}

	exprSQL, err := compileExpr(node.Expr, input.Schema)
	if err != nil {
		ctx.addDiag(node.DiagAnchor, fmt.Sprintf("invalid expression: %v", err), "error")
		return compiledRel{}, fmt.Errorf("DeriveColumn expression compilation failed (anchor: %s): %w", node.DiagAnchor, err)
	}

	cteName := ctx.nextCTEName()
	sql := fmt.Sprintf("%s AS (\n  SELECT *, %s AS %s\n  FROM %s\n)", cteName, exprSQL, node.ColumnName, input.CTEName)
	ctx.ctes = append(ctx.ctes, sql)
	ctx.diagMap[cteName] = node.DiagAnchor

	// Output schema = input schema + the new derived column.
	// DataType of the derived column is unknown at this stage (no type inference in Phase 2).
	// Use an empty DataType as a placeholder.
	outputSchema := append(append(rel.Schema{}, input.Schema...), rel.Column{
		Name: node.ColumnName,
		Type: "",
	})

	return compiledRel{
		CTEName:    cteName,
		SQL:        sql,
		Schema:     outputSchema,
		DiagAnchor: node.DiagAnchor,
	}, nil
}

// compileExpr converts an Expr node into a SQL fragment string.
// schema is used to validate column references.
func compileExpr(e expr.Expr, schema rel.Schema) (string, error) {
	switch ex := e.(type) {
	case expr.Literal:
		return compileLiteral(ex)
	case expr.ColumnRef:
		return compileColumnRef(ex, schema)
	case expr.BinaryOp:
		return compileBinaryOp(ex, schema)
	case expr.FuncCall:
		return compileFuncCall(ex, schema)
	default:
		return "", fmt.Errorf("unsupported Expr type: %T", e)
	}
}

func compileLiteral(e expr.Literal) (string, error) {
	switch e.DataType {
	case expr.DataTypeString:
		// Single-quote the value and escape internal single quotes.
		s := fmt.Sprintf("%v", e.Value)
		s = strings.ReplaceAll(s, "'", "''")
		return fmt.Sprintf("'%s'", s), nil
	case expr.DataTypeInt, expr.DataTypeFloat:
		return fmt.Sprintf("%v", e.Value), nil
	case expr.DataTypeBool:
		b, ok := e.Value.(bool)
		if !ok {
			return "", fmt.Errorf("Literal DataType is bool but Value is %T", e.Value)
		}
		if b {
			return "TRUE", nil
		}
		return "FALSE", nil
	default:
		return "", fmt.Errorf("unsupported DataType: %q", e.DataType)
	}
}

func compileColumnRef(e expr.ColumnRef, schema rel.Schema) (string, error) {
	if err := resolveColumn(e.Name, schema); err != nil {
		return "", err
	}
	return e.Name, nil
}

func compileBinaryOp(e expr.BinaryOp, schema rel.Schema) (string, error) {
	left, err := compileExpr(e.Left, schema)
	if err != nil {
		return "", err
	}
	right, err := compileExpr(e.Right, schema)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("(%s %s %s)", left, string(e.Op), right), nil
}

func compileFuncCall(e expr.FuncCall, schema rel.Schema) (string, error) {
	args := make([]string, len(e.Args))
	for i, arg := range e.Args {
		s, err := compileExpr(arg, schema)
		if err != nil {
			return "", fmt.Errorf("func %q arg[%d]: %w", e.Name, i, err)
		}
		args[i] = s
	}
	return fmt.Sprintf("%s(%s)", e.Name, strings.Join(args, ", ")), nil
}

// resolveColumn checks that name exists in schema using case-sensitive exact match.
// Column resolution logic is centralised here; do not duplicate elsewhere.
func resolveColumn(name string, schema rel.Schema) error {
	for _, col := range schema {
		if col.Name == name {
			return nil
		}
	}
	return fmt.Errorf("column %q not found in schema", name)
}