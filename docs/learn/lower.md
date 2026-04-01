# lower（AST → IR変換）

→ [全体俯瞰に戻る](./overview.md)

## 一言で言うと

**ASTのツリーをIRのツリーに翻訳する処理。**
入力: `ast.ASTNode`、出力: `rel.RelNode`。

## どこにある

`internal/domain/ir/lower/lower.go`

## 変換ルール

lowerはASTノードのTypeとValueを見て、対応するIRノードに変換する。

| ASTノード | → | IRノード |
|---|---|---|
| `table_query` | → | `rel.TableScan` |
| `formula / WHERE` | → | `rel.Filter` |
| `formula / SELECT` | → | `rel.Project` |
| `formula / CROSS_JOIN` | → | `rel.Join{Kind: CROSS}` |
| `formula / JOIN` | → | `rel.Join{Kind: INNER}` |
| `formula / LEFT_JOIN` | → | `rel.Join{Kind: LEFT}` |
| `formula / GROUP_BY` | → | `rel.Aggregate` |
| それ以外のformula / operator / variable / literal | → | `rel.DeriveColumn`（式をwrap） |
| `const` | → | `expr.Literal`（ConstResolverで解決） |

## 具体例

### WHERE(motor, model = "A111")

```
【AST】
ASTNode{Type: formula, Value: "WHERE"}
  ├── ASTNode{Type: table_query, Value: "motor"}
  └── ASTNode{Type: operator, Value: "="}
        ├── ASTNode{Type: variable, Value: "model"}
        └── ASTNode{Type: literal,  Value: "A111"}

        ↓ lower

【IR】
rel.Filter{
    Input:     rel.TableScan{Table: {Name: "motor"}},
    Predicate: expr.BinaryOp{
        Op:    "=",
        Left:  expr.ColumnRef{Name: "model"},
        Right: expr.Literal{Value: "A111", DataType: string},
    },
}
```

### rpm * coef（スカラー式）

スカラー式はそのままではIRに乗らない（IRはテーブル操作の言語）。
lowerはスカラー式を`DeriveColumn`でwrapして、テーブル操作として表現する。

```
【AST】
ASTNode{Type: operator, Value: "*"}
  ├── ASTNode{Type: variable, Value: "rpm"}
  └── ASTNode{Type: variable, Value: "coef"}

        ↓ lower

【IR】
rel.DeriveColumn{
    ColumnName: "__ff_result",   ← 内部予約列名
    Expr: expr.BinaryOp{
        Op:    "*",
        Left:  expr.ColumnRef{Name: "rpm"},
        Right: expr.ColumnRef{Name: "coef"},
    },
    Input: ???  ← 別途TableScanが必要
}
```

`__ff_`プレフィックスはformuflow内部予約。ユーザーには見えない。

### CROSS_JOIN(rpm_table, coef_table)

```
【AST】
ASTNode{Type: formula, Value: "CROSS_JOIN"}
  ├── ASTNode{Type: table_query, Value: "rpm_table"}
  └── ASTNode{Type: table_query, Value: "coef_table"}

        ↓ lower

【IR】
rel.Join{
    Kind:  JoinCross,
    Left:  rel.TableScan{Table: {Name: "rpm_table"}},
    Right: rel.TableScan{Table: {Name: "coef_table"}},
    On:    nil,
}
```

## lowerは検証しない

- 型チェック、列存在チェック、関数引数チェック → App層の責務
- lowerはASTの形（childrenの数・種類）だけチェックする
- 列が本当に存在するかの検証はsqlgenのコンパイル時に行われる