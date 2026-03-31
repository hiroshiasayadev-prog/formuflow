package lower

import (
	"fmt"
	"strconv"

	"github.com/your-module/internal/domain/ir/ast"
	"github.com/your-module/internal/domain/ir/expr"
	"github.com/your-module/internal/domain/ir/rel"
)

// validOperators is the set of binary operators permitted in Phase 2.
var validOperators = map[string]bool{
	"+": true, "-": true, "*": true, "/": true,
	">": true, "<": true, "=": true,
}

// Lower converts a FormulaNode AST into a rel.RelNode tree.
// The output is always a rel.RelNode; pure scalar expressions are wrapped in DeriveColumn.
// consts is used to resolve NodeTypeConst names to typed values.
func Lower(root *ast.ASTNode, consts ast.ConstResolver) (rel.RelNode, error) {
	ctx := &lowerContext{consts: consts}
	return ctx.lowerNode(root)
}

// lowerContext holds mutable state for a single Lower invocation.
type lowerContext struct {
	consts  ast.ConstResolver
	counter int // counter for __ff_result_N column names
}

// resultColumnName generates a unique reserved column name for DeriveColumn wraps.
// The __ff_ prefix is reserved for formuflow internal use.
func (ctx *lowerContext) resultColumnName() string {
	ctx.counter++
	if ctx.counter == 1 {
		return "__ff_result"
	}
	return fmt.Sprintf("__ff_result_%d", ctx.counter)
}

// lowerNode converts an ASTNode into a rel.RelNode.
// Relation-returning builtins (WHERE, SELECT) produce their corresponding RelNodes.
// All other nodes are lowered as scalar expressions and wrapped in DeriveColumn.
func (ctx *lowerContext) lowerNode(node *ast.ASTNode) (rel.RelNode, error) {
	switch node.Type {
	case ast.NodeTypeTableQuery:
		return ctx.lowerTableQuery(node)
	case ast.NodeTypeFormula:
		switch node.Value {
		case "WHERE":
			return ctx.lowerWhere(node)
		case "SELECT":
			return ctx.lowerSelect(node)
		}
		fallthrough
	default:
		// Pure scalar expression — lower as expr and wrap in DeriveColumn.
		e, err := ctx.lowerExpr(node)
		if err != nil {
			return nil, err
		}
		return rel.DeriveColumn{
			DiagAnchor: node.ID,
			Expr:       e,
			ColumnName: ctx.resultColumnName(),
		}, nil
	}
}

// lowerTableQuery converts a NodeTypeTableQuery into a rel.TableScan.
func (ctx *lowerContext) lowerTableQuery(node *ast.ASTNode) (rel.RelNode, error) {
	if node.Value == "" {
		return nil, fmt.Errorf("lower: NodeTypeTableQuery has empty table name (anchor: %s)", node.ID)
	}
	return rel.TableScan{
		DiagAnchor: node.ID,
		Table:      rel.TableRef{Name: node.Value},
	}, nil
}

// lowerWhere converts a WHERE builtin call into a rel.Filter.
//
// Shape requirements:
//   - exactly 2 children
//   - Children[0] must be NodeTypeTableQuery
func (ctx *lowerContext) lowerWhere(node *ast.ASTNode) (rel.RelNode, error) {
	if len(node.Children) != 2 {
		return nil, fmt.Errorf("lower: WHERE requires exactly 2 arguments, got %d (anchor: %s)", len(node.Children), node.ID)
	}
	if node.Children[0].Type != ast.NodeTypeTableQuery {
		return nil, fmt.Errorf("lower: WHERE first argument must be a table query, got %s (anchor: %s)", node.Children[0].Type, node.ID)
	}

	table, err := ctx.lowerNode(node.Children[0])
	if err != nil {
		return nil, err
	}

	pred, err := ctx.lowerExpr(node.Children[1])
	if err != nil {
		return nil, err
	}

	return rel.Filter{
		DiagAnchor: node.ID,
		Input:      table,
		Predicate:  pred,
	}, nil
}

// lowerSelect converts a SELECT builtin call into a rel.Project.
//
// Shape requirements:
//   - at least 2 children
//   - Children[0] must be NodeTypeTableQuery
//   - Children[1:] must all be NodeTypeLiteral with DataTypeString
func (ctx *lowerContext) lowerSelect(node *ast.ASTNode) (rel.RelNode, error) {
	if len(node.Children) < 2 {
		return nil, fmt.Errorf("lower: SELECT requires at least 2 arguments, got %d (anchor: %s)", len(node.Children), node.ID)
	}
	if node.Children[0].Type != ast.NodeTypeTableQuery {
		return nil, fmt.Errorf("lower: SELECT first argument must be a table query, got %s (anchor: %s)", node.Children[0].Type, node.ID)
	}

	for i, child := range node.Children[1:] {
		if child.Type != ast.NodeTypeLiteral || child.DataType != expr.DataTypeString {
			return nil, fmt.Errorf("lower: SELECT column argument %d must be a string literal, got type=%s datatype=%s (anchor: %s)", i+1, child.Type, child.DataType, node.ID)
		}
	}

	table, err := ctx.lowerNode(node.Children[0])
	if err != nil {
		return nil, err
	}

	columns := make([]string, len(node.Children)-1)
	for i, child := range node.Children[1:] {
		columns[i] = child.Value
	}

	return rel.Project{
		DiagAnchor: node.ID,
		Input:      table,
		Columns:    columns,
	}, nil
}

// lowerExpr converts an ASTNode into an expr.Expr (scalar expression).
// This is called for predicate expressions and non-relation-returning formula bodies.
func (ctx *lowerContext) lowerExpr(node *ast.ASTNode) (expr.Expr, error) {
	switch node.Type {
	case ast.NodeTypeLiteral:
		return lowerLiteral(node)
	case ast.NodeTypeConst:
		return ctx.lowerConst(node)
	case ast.NodeTypeVariable:
		return expr.ColumnRef{Name: node.Value}, nil
	case ast.NodeTypeOperator:
		return ctx.lowerOperator(node)
	case ast.NodeTypeFormula:
		return ctx.lowerFuncCall(node)
	default:
		return nil, fmt.Errorf("lower: unsupported node type in expression context: %s (anchor: %s)", node.Type, node.ID)
	}
}

// lowerLiteral converts a NodeTypeLiteral into an expr.Literal.
// Value is always stored as string in ASTNode and converted here to the appropriate Go type.
func lowerLiteral(node *ast.ASTNode) (expr.Expr, error) {
	v, err := convertValue(node.Value, node.DataType)
	if err != nil {
		return nil, fmt.Errorf("lower: literal conversion failed (anchor: %s): %w", node.ID, err)
	}
	return expr.Literal{Value: v, DataType: node.DataType}, nil
}

// lowerConst resolves a NodeTypeConst name via ConstResolver and embeds it as expr.Literal.
func (ctx *lowerContext) lowerConst(node *ast.ASTNode) (expr.Expr, error) {
	v, dt, err := ctx.consts.Resolve(node.Value)
	if err != nil {
		return nil, fmt.Errorf("lower: unresolved constant %q (anchor: %s): %w", node.Value, node.ID, err)
	}
	return expr.Literal{Value: v, DataType: dt}, nil
}

// lowerOperator converts a NodeTypeOperator into an expr.BinaryOp.
//
// Shape requirements:
//   - exactly 2 children
//   - Value must be one of: + - * / > < =
func (ctx *lowerContext) lowerOperator(node *ast.ASTNode) (expr.Expr, error) {
	if len(node.Children) != 2 {
		return nil, fmt.Errorf("lower: operator %q requires exactly 2 children, got %d (anchor: %s)", node.Value, len(node.Children), node.ID)
	}
	if !validOperators[node.Value] {
		return nil, fmt.Errorf("lower: unsupported operator %q (anchor: %s)", node.Value, node.ID)
	}

	left, err := ctx.lowerExpr(node.Children[0])
	if err != nil {
		return nil, err
	}
	right, err := ctx.lowerExpr(node.Children[1])
	if err != nil {
		return nil, err
	}

	return expr.BinaryOp{
		Op:    expr.BinaryOpKind(node.Value),
		Left:  left,
		Right: right,
	}, nil
}

// lowerFuncCall converts a non-builtin NodeTypeFormula into an expr.FuncCall.
// The function name and arguments are passed through to DuckDB as-is.
func (ctx *lowerContext) lowerFuncCall(node *ast.ASTNode) (expr.Expr, error) {
	args := make([]expr.Expr, len(node.Children))
	for i, child := range node.Children {
		arg, err := ctx.lowerExpr(child)
		if err != nil {
			return nil, err
		}
		args[i] = arg
	}
	return expr.FuncCall{Name: node.Value, Args: args}, nil
}

// convertValue converts a string representation of a value to the appropriate Go type
// based on DataType. This is the responsibility of lower, not the compiler.
func convertValue(s string, dt expr.DataType) (any, error) {
	switch dt {
	case expr.DataTypeInt:
		v, err := strconv.Atoi(s)
		if err != nil {
			return nil, fmt.Errorf("cannot convert %q to int: %w", s, err)
		}
		return v, nil
	case expr.DataTypeFloat:
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return nil, fmt.Errorf("cannot convert %q to float: %w", s, err)
		}
		return v, nil
	case expr.DataTypeBool:
		v, err := strconv.ParseBool(s)
		if err != nil {
			return nil, fmt.Errorf("cannot convert %q to bool: %w", s, err)
		}
		return v, nil
	case expr.DataTypeString:
		return s, nil
	default:
		return nil, fmt.Errorf("unsupported DataType: %q", dt)
	}
}