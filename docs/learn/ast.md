# AST（抽象構文木）

→ [全体俯瞰に戻る](./overview.md)

## 一言で言うと

**ユーザーが書いた式をツリーで表現したもの。**
App層が作って、lowerに渡す。

## どこにある

`internal/domain/ir/ast/ast.go`

## 何が入っているか

### ASTNode — 式の1ノード

```go
type ASTNode struct {
    ID       string        // DiagAnchorに使う（エラー追跡用）
    Type     NodeType      // このノードの種類
    Value    string        // リテラル値 / 演算子記号 / 関数名 / 変数名
    Children []*ASTNode    // 子ノード
    DataType expr.DataType // 値の型
}
```

### NodeType — ノードの種類

| NodeType | 意味 | Valueに入るもの |
|---|---|---|
| `literal` | 式中の定数値 | `"42"`, `"3.14"`, `"true"` |
| `const` | 名前付き定数（実行時に解決） | `"model_name"` |
| `operator` | 二項演算子 | `"+"`, `"-"`, `">"`, `"="` など |
| `variable` | 列参照 | `"age"`, `"rpm"` |
| `formula` | 関数呼び出し | `"WHERE"`, `"SUM"`, `"CROSS_JOIN"` など |
| `table_query` | DBテーブル参照 | `"motor_specs"`, `"users"` |

### ConstResolver — 名前付き定数の解決

```go
type ConstResolver interface {
    Resolve(name string) (any, expr.DataType, error)
}
```

`const`ノードはValueに定数名が入っているだけで、実際の値は持っていない。
lower時にConstResolver経由で値を取得してexpr.Literalに変換する。

## 具体例

ユーザーが `age * 10` という式を定義した場合：

```
ASTNode{Type: operator, Value: "*"}
  ├── ASTNode{Type: variable, Value: "age"}
  └── ASTNode{Type: literal,  Value: "10", DataType: int}
```

`WHERE(motor, model = "A111")` の場合：

```
ASTNode{Type: formula, Value: "WHERE"}
  ├── ASTNode{Type: table_query, Value: "motor"}
  └── ASTNode{Type: operator, Value: "="}
        ├── ASTNode{Type: variable, Value: "model"}
        └── ASTNode{Type: literal,  Value: "A111", DataType: string}
```

## ASTとIRの違い

ASTは「式の言語」。IRは「テーブル操作の言語」。

| | AST | IR（rel） |
|---|---|---|
| 表現するもの | ユーザーの式 | テーブル変換の構造 |
| 葉ノード | literal, variable | TableScan |
| 変換ノード | operator, formula | Filter, Project, DeriveColumn |
| 作るもの | App層 | lower |
| 消費するもの | lower | sqlgen |

lowerがASTをIRに翻訳する。→ [lower解説](./lower.md)