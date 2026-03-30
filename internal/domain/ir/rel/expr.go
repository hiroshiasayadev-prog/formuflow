package expr

// Expr is the interface for all expression nodes in the IR.
// Expressions represent scalar computations that are compiled into SQL fragments.
// Evaluation is delegated to DuckDB; the IR itself does not evaluate expressions.
type Expr interface {
	exprNode()
}

// Literal represents a constant value.
// Value is held as any because the type is already resolved by the App layer;
// the IR does not re-infer types. At compile time, Value is rendered as a SQL
// literal according to DataType.
type Literal struct {
	Value    any
	DataType DataType
}

// ColumnRef represents a reference to a column by name.
// Table is reserved for future JOIN support and is unused in Phase 2.
type ColumnRef struct {
	Name  string
	Table string // optional; reserved for future JOIN / alias support
}

// BinaryOp represents a binary operation between two expressions.
type BinaryOp struct {
	Op    BinaryOpKind
	Left  Expr
	Right Expr
}

// FuncCall represents a function invocation.
// Argument type checking and function registry lookups are the responsibility
// of the App layer. The IR and compiler pass the name and args through to SQL as-is.
type FuncCall struct {
	Name string
	Args []Expr
}

func (Literal) exprNode()   {}
func (ColumnRef) exprNode() {}
func (BinaryOp) exprNode()  {}
func (FuncCall) exprNode()  {}

// BinaryOpKind enumerates the supported binary operators.
type BinaryOpKind string

const (
	OpAdd BinaryOpKind = "+"
	OpSub BinaryOpKind = "-"
	OpMul BinaryOpKind = "*"
	OpDiv BinaryOpKind = "/"
	OpGT  BinaryOpKind = ">"
	OpLT  BinaryOpKind = "<"
	OpEQ  BinaryOpKind = "="
)

// DataType enumerates the scalar types that the IR recognises.
// These correspond to the Scalar type variants defined in the domain model.
type DataType string

const (
	DataTypeInt    DataType = "int"
	DataTypeFloat  DataType = "float"
	DataTypeString DataType = "string"
	DataTypeBool   DataType = "bool"
)