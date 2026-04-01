package rel

import "github.com/hiroshiasayadev-prog/formuflow/internal/domain/ir/expr"

// RelNode is the interface for all relational IR nodes.
// Each node represents a table-transforming operation and forms a tree structure
// where each node holds its input as a child RelNode.
// The tree is evaluated bottom-up (child before parent) during compilation.
type RelNode interface {
	relNode()
}

// Column describes a single column in a schema.
// Used as compile-time metadata for column resolution and validation.
// SourceID and SourceName are reserved for future JOIN support.
type Column struct {
	Name string
	Type expr.DataType
	// future: SourceID, SourceName for JOIN support
}

// Schema is the ordered list of columns output by a RelNode.
// It is computed during compilation and propagated from child to parent.
// Schema is internal to the compiler and is not stored in the IR nodes themselves.
type Schema []Column

// SchemaProvider resolves a table name to its schema at compile time.
// In Phase 2, this is backed by the database or metadata store.
type SchemaProvider interface {
	GetSchema(tableName string) (Schema, error)
}

// TableScan is a leaf node that reads from a named table.
// It has no input node; schema is obtained via SchemaProvider at compile time.
type TableScan struct {
	DiagAnchor string
	Table      TableRef
}

// TableRef identifies the target table of a TableScan.
// Additional fields (Schema, Catalog, SourceKind, Alias) are reserved for future use.
type TableRef struct {
	Name string
	// future: Schema, Catalog, SourceKind, Alias
}

// Filter retains only the rows from Input that satisfy Predicate.
// Corresponds to a SQL WHERE clause.
type Filter struct {
	DiagAnchor string
	Input      RelNode
	Predicate  expr.Expr
}

// Project selects and reorders a subset of columns from Input.
// Only columns already present in the input schema may be listed.
// Renaming, expression evaluation, and constant columns are not supported;
// use DeriveColumn for those cases.
type Project struct {
	DiagAnchor string
	Input      RelNode
	Columns    []string
}

// DeriveColumn appends a single computed column to the output of Input.
// The column is named ColumnName and its value is computed by Expr.
//
// Rules:
//   - Expr may only reference columns present in Input's output schema.
//   - ColumnName must not duplicate any column already in Input's output schema.
//   - To derive multiple columns where one depends on another, nest DeriveColumn nodes:
//
//     DeriveColumn(ColumnName="b", Expr=a*2)
//     └── DeriveColumn(ColumnName="a", Expr=x+1)
//     └── TableScan("users")
type DeriveColumn struct {
	DiagAnchor string
	Input      RelNode
	ColumnName string
	Expr       expr.Expr
}

func (TableScan) relNode()    {}
func (Filter) relNode()       {}
func (Project) relNode()      {}
func (DeriveColumn) relNode() {}


// JoinKind enumerates the supported join types.
type JoinKind string

const (
    JoinCross JoinKind = "CROSS"
    JoinInner JoinKind = "INNER"
    JoinLeft  JoinKind = "LEFT"
)

// Join combines two input relations into one.
// It is the only RelNode with two inputs (Left and Right).
//
// Output schema: all columns from Left followed by all columns from Right.
// Column name conflicts between Left and Right are treated as a compile error.
//
// Rules:
//   - JoinCross: On must be nil. Output is the cartesian product (Left × Right rows).
//   - JoinInner / JoinLeft: On must be a non-nil predicate expr.
type Join struct {
    DiagAnchor string
    Kind       JoinKind
    Left       RelNode
    Right      RelNode
    On         expr.Expr // nil for CROSS; required for INNER / LEFT
}

func (Join) relNode() {}

// AggFunc enumerates the supported aggregate functions.
type AggFunc string

const (
    AggSum   AggFunc = "SUM"
    AggCount AggFunc = "COUNT"
    AggAvg   AggFunc = "AVG"
    AggMin   AggFunc = "MIN"
    AggMax   AggFunc = "MAX"
)

// AggExpr describes a single aggregation: Func(Col) AS Alias.
type AggExpr struct {
    Func  AggFunc
    Col   string // column name to aggregate
    Alias string // output column name
}

// Aggregate groups rows by GroupBy columns and applies aggregate functions.
// Corresponds to SQL GROUP BY + aggregate functions.
//
// Output schema: GroupBy columns followed by Aggs alias columns.
// If GroupBy is empty, the entire input is collapsed into a single row.
//
// Rules:
//   - Aggs must not be empty.
//   - Each Col in Aggs must exist in the input schema.
//   - Each column in GroupBy must exist in the input schema.
type Aggregate struct {
    DiagAnchor string
    Input      RelNode
    GroupBy    []string
    Aggs       []AggExpr
}

func (Aggregate) relNode() {}