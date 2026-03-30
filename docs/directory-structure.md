# formuflow backend directory structure

## decisions

- Clean architecture: domain / usecase / adapter / infrastructure
- IR lives in `domain/ir/` — it is the internal representation of the domain's computation model
- SQL compilation lives in `adapter/sqlgen/` — outward-facing transformation from IR to SQL string

## package layout

```
internal/
  domain/
    ir/
      expr/
        expr.go       -- Expr interface, Literal, ColumnRef, BinaryOp, FuncCall, DataType
      rel/
        rel.go        -- RelNode interface, Schema, SchemaProvider, TableScan, Filter, Project, DeriveColumn
  adapter/
    sqlgen/
      compile.go      -- RelNode → SQL string + Diagnostics (CompileResult, Diagnostic, compileRel, compileExpr)
```

## dependency direction

```
expr  ←  rel  ←  sqlgen
```

- `rel` imports `expr` (e.g. Filter.Predicate is an expr.Expr)
- `sqlgen` imports both `rel` and `expr`
- `domain/ir/expr` and `domain/ir/rel` have no dependency on adapter layer

## what goes where

| concept | package | notes |
|---|---|---|
| Expr interface + nodes | `domain/ir/expr` | pure value types, no evaluation |
| RelNode interface + nodes | `domain/ir/rel` | tree structure, no compilation logic |
| Schema, SchemaProvider | `domain/ir/rel` | compile-time metadata, internal to compiler |
| SQL compilation | `adapter/sqlgen` | CompileResult, Diagnostic, CTE generation |
| App layer nodes (Component) | `domain/component/` | separate from IR; lowered into IR before compilation |

## notes

- IR nodes do not evaluate themselves; evaluation is delegated to DuckDB
- Schema is computed during compilation and propagated child → parent; it is not stored in IR nodes
- `ColumnRef.Table` and `TableRef` extra fields are reserved for future JOIN support, unused in Phase 2
- Replace `github.com/your-module` in import paths with the actual module name from `go.mod`