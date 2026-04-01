# expr（スカラー式）

→ [全体俯瞰に戻る](./overview.md)

## 一言で言うと

**スカラー計算を表すノードの定義。**
IRの部品として rel の中に埋め込まれる。自分では評価しない。

## どこにある

`internal/domain/ir/expr/expr.go`

## ASTのexprとの違い

紛らわしいが別物。

| | ast.ASTNode | expr.Expr |
|---|---|---|
| 作るもの | App層 | lower |
| 消費するもの | lower | sqlgen |
| 表現するもの | ユーザーの式（生） | IR内のスカラー式（変換済み） |

ASTはlowerへの入力、exprはlowerの出力（の一部）。

## ノード一覧

### Literal — 定数値

```go
type Literal struct {
    Value    any      // Go の値（int / float64 / string / bool）
    DataType DataType // 型情報
}
```

例: `42`、`"A111"`、`true`

App層で型が確定しているのでIR側では型推論しない。
sqlgenでDataTypeに応じてSQL literalに変換される。

```
Literal{Value: 42, DataType: int}  →  SQL: 42
Literal{Value: "A111", DataType: string}  →  SQL: 'A111'
Literal{Value: true, DataType: bool}  →  SQL: TRUE
```

### ColumnRef — 列参照

```go
type ColumnRef struct {
    Name  string
    Table string // 将来のJOIN用に予約。Phase 2/3では空
}
```

例: `age`、`rpm`

sqlgenでinput schemaに列が存在するか検証される。

### BinaryOp — 二項演算

```go
type BinaryOp struct {
    Op    BinaryOpKind // + - * / > < =
    Left  Expr
    Right Expr
}
```

例: `age * 10`、`model = "A111"`

```
BinaryOp{Op: "*", Left: ColumnRef{Name:"age"}, Right: Literal{Value:10}}
→ SQL: (age * 10)
```

### FuncCall — 関数呼び出し

```go
type FuncCall struct {
    Name string
    Args []Expr
}
```

例: `SUM(salary)`、`ROUND(price, 2)`

関数の型チェックはApp層の責務。IRとsqlgenは名前と引数をそのままDuckDBに渡す。

## どこに埋め込まれるか

exprノードは単独では存在せず、relノードの中に埋め込まれる。

```
rel.Filter{
    Predicate: expr.BinaryOp{   ← ここ
        Op:    ">",
        Left:  expr.ColumnRef{Name: "age"},
        Right: expr.Literal{Value: 30, DataType: int},
    }
}

rel.DeriveColumn{
    Expr: expr.BinaryOp{        ← ここ
        Op:    "*",
        Left:  expr.ColumnRef{Name: "age"},
        Right: expr.Literal{Value: 10, DataType: int},
    }
}
```

Phase 2/3で式を持てるrelノード: Filter（Predicate）、DeriveColumn（Expr）、Join（On）。