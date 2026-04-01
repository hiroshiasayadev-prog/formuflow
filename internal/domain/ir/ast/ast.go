package ast

import "github.com/hiroshiasayadev-prog/formuflow/internal/domain/ir/expr"

// NodeType represents the type of an AST node.
type NodeType string

const (
	NodeTypeLiteral    NodeType = "literal"     // formula-local literal value
	NodeTypeConst      NodeType = "const"       // named constant; resolved via ConstResolver at lower time
	NodeTypeOperator   NodeType = "operator"    // binary operator: + - * / > < =
	NodeTypeVariable   NodeType = "variable"    // column reference → expr.ColumnRef
	NodeTypeFormula    NodeType = "formula"     // function call; dispatched by name in lower
	NodeTypeTableQuery NodeType = "table_query" // DB table reference → rel.TableScan
)

// ASTNode is a node in the formula expression tree.
// It is produced by the App layer and consumed by lower.
// SQL generation cache, position info, and aggregate flags from the prototype are omitted;
// those concerns belong to the compiler and App layer respectively.
type ASTNode struct {
	ID       string      // used as DiagAnchor in IR nodes
	Type     NodeType
	Value    string      // literal value (string repr) / operator symbol / function name / variable name
	Children []*ASTNode
	DataType expr.DataType
}

// ConstResolver resolves a named constant to its typed Go value.
// It is supplied by the App layer and passed into lower.
// Resolve must return a Go value consistent with the returned DataType.
type ConstResolver interface {
	Resolve(name string) (any, expr.DataType, error)
}