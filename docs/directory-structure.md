# formuflow backend directory structure

## decisions

- Clean architecture: domain / usecase / adapter / infrastructure
- IR lives in `domain/ir/` — it is the internal representation of the domain's computation model
- Lower lives in `domain/ir/lower/` — domain logic, not an adapter concern
- SQL compilation lives in `adapter/sqlgen/` — outward-facing transformation from IR to SQL string

## package layout

```
internal/
  domain/
    ir/
      ast/
        ast.go        -- ASTNode, NodeType, ConstResolver
      expr/
        expr.go       -- Expr interface, Literal, ColumnRef, BinaryOp, FuncCall, DataType
      rel/
        rel.go        -- RelNode interface, Schema, SchemaProvider, TableScan, Filter, Project, DeriveColumn
      lower/
        lower.go      -- ASTNode → rel.RelNode (Lower entry point, lowerNode, lowerExpr)
  adapter/
    sqlgen/
      compile.go      -- rel.RelNode → SQL string + Diagnostics (CompileResult, Diagnostic, compileRel, compileExpr)
```

## dependency direction

```
ast  ←  expr  ←  rel  ←  lower
                  rel  ←  sqlgen
                 expr  ←  sqlgen
```

- `ast` imports `expr` (ASTNode.DataType is expr.DataType)
- `rel` imports `expr` (e.g. Filter.Predicate is an expr.Expr)
- `lower` imports `ast`, `expr`, `rel`
- `sqlgen` imports `rel` and `expr`
- `domain/ir/expr` and `domain/ir/rel` have no dependency on adapter layer

## what goes where

| concept | package | notes |
|---|---|---|
| ASTNode, NodeType, ConstResolver | `domain/ir/ast` | formula expression tree; produced by App layer |
| Expr interface + nodes | `domain/ir/expr` | pure value types, no evaluation |
| RelNode interface + nodes | `domain/ir/rel` | tree structure, no compilation logic |
| Schema, SchemaProvider | `domain/ir/rel` | compile-time metadata, internal to compiler |
| ASTNode → RelNode conversion | `domain/ir/lower` | lower logic; shape validation; literal conversion |
| SQL compilation | `adapter/sqlgen` | CompileResult, Diagnostic, CTE generation |
| App layer nodes (Component) | `domain/component/` | separate from IR; lowered into IR before compilation |

## builtin function names (Phase 2)

SQL-inspired names used in ASTNode.Value for relation-returning builtins:

| builtin | lower result | notes |
|---|---|---|
| `WHERE` | rel.Filter | arg[0]: table query, arg[1]: predicate |
| `SELECT` | rel.Project | arg[0]: table query, arg[1..]: string literal column names |

All other function names become expr.FuncCall and are passed through to DuckDB.

## notes

- IR nodes do not evaluate themselves; evaluation is delegated to DuckDB
- Schema is computed during compilation and propagated child → parent; it is not stored in IR nodes
- `ColumnRef.Table` and `TableRef` extra fields are reserved for future JOIN support, unused in Phase 2
- `__ff_` column name prefix is reserved for formuflow internal use (e.g. `__ff_result` for wrapped scalar expressions)
- Replace `github.com/your-module` in import paths with the actual module name from `go.mod`