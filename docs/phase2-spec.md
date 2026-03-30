# formuflow Phase 2 仕様書

## アーキテクチャ概要

```
App Node (DB保存)
  → App層でバリデーション
  → lower
  → IR (rel + expr)
  → compile
  → SQL文字列 + Diag
  → DuckDB実行
```

---

## パッケージ構成と依存関係

```
expr/    式の木構造
rel/     テーブル変換の木構造（IRノード定義）
sql/     RelNode → SQL文字列のコンパイラ
```

```
expr  ←  rel  ←  sql
```

---

## expr/ 型定義

```go
type Expr interface {
    exprNode()
}

// リテラル値
// Value は any で保持する。App層で型が確定しているので IR 側で再推論しない。
// compile時に DataType に応じて SQL literal 化する。
type Literal struct {
    Value    any
    DataType DataType
}

// 列参照
// Table は optional（JOIN等の将来拡張用）。Phase 2 では空でよい。
type ColumnRef struct {
    Name  string
    Table string // optional、Phase 2 では未使用
}

// 二項演算子
type BinaryOp struct {
    Op    BinaryOpKind
    Left  Expr
    Right Expr
}

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

// 関数呼び出し
// 関数定義テーブルと引数型チェックは App層の責務。
// IR・compile側では関数名と引数をそのまま SQL に落とす。
type FuncCall struct {
    Name string
    Args []Expr
}

type DataType string

const (
    DataTypeInt    DataType = "int"
    DataTypeFloat  DataType = "float"
    DataTypeString DataType = "string"
    DataTypeBool   DataType = "bool"
)
```

### 式の責務分担

- `Expr` は式の構文木であり、自身では評価しない
- `Expr` は `sql/` で SQL 断片へコンパイルされ、実際の評価は DuckDB が行う
- Phase 2 において、式を保持できる RelNode は `Filter` と `DeriveColumn` のみ
- `Literal`, `ColumnRef` に DiagAnchor は持たせない（App層バリデーションを信頼）

---

## rel/ 型定義

### Schema

Schema は compile 時の列解決に使う metadata として扱う。表示用の一覧ではない。

```go
// Phase 2 では Name と Type のみ。
// 将来 JOIN 対応時に SourceID / SourceName 等を追加できるよう拡張余地を残す。
type Column struct {
    Name string
    Type DataType
}

type Schema []Column
```

### SchemaProvider

TableScan の Schema は compile 時に外部から提供される。
compile は `SchemaProvider` 経由で TableRef.Name から Schema を取得する。

```go
type SchemaProvider interface {
    GetSchema(tableName string) (Schema, error)
}
```

Phase 2 では DB または metadata から提供される前提とし、compile に渡す。

### RelNode

```go
type RelNode interface {
    relNode()
}

type TableScan struct {
    DiagAnchor string
    Table      TableRef
}

type TableRef struct {
    Name string
    // future: Schema, Catalog, SourceKind, Alias
}

type Filter struct {
    DiagAnchor string
    Input      RelNode
    Predicate  Expr
}

// Project は input に既に存在する列の選択と順序指定のみを行う。
// rename, 式評価, 定数列追加はサポートしない。
// 新規列追加は DeriveColumn で表現する。
type Project struct {
    DiagAnchor string
    Input      RelNode
    Columns    []string
}

type DeriveColumn struct {
    DiagAnchor string
    Input      RelNode
    ColumnName string
    Expr       Expr
}
```

### DeriveColumn 参照ルール

- `Expr` が参照できるのは `Input` の出力スキーマに存在する列のみ
- 自ノードが追加する `ColumnName` は参照不可
- 派生列に依存したい場合は、その列を生成した DeriveColumn を子側に置く
- `ColumnName` が input schema の既存列と重複する場合はエラーとする（上書き禁止）

```
// b が a を参照したい場合
DeriveColumn(ColumnName="b", Expr=a*2)
  └── DeriveColumn(ColumnName="a", Expr=x+1)
        └── TableScan("users")
```

木は下から上に評価される（子→親の順）。

---

## sql/ 型定義

```go
type CompileResult struct {
    SQL   string
    Diags []Diagnostic
    // future: Lineage, NodeSpanMap
}

type Diagnostic struct {
    DiagAnchor string
    Message    string
    Severity   string // "error", "warning"
}
```

### compile内部表現

compile の各ステップは以下の内部型を返す。
将来 JOIN 等で情報が増えた場合はこの struct を拡張する。

```go
type compiledRel struct {
    CTEName    string
    SQL        string  // このノードのCTE本体
    Schema     Schema  // このノードの output schema
    DiagAnchor string
}
```

### compile関数のシグネチャ（概念）

```go
// RelNode を再帰的にコンパイルし、compiledRel を返す
func compileRel(node RelNode, provider SchemaProvider, cteSeq *int, ctes *[]string, diagMap map[string]string) (compiledRel, error)

// Expr を SQL 断片文字列に変換する
func compileExpr(e Expr, schema Schema) (string, error)
```

---

## SQL生成方針

### CTE採用

サブクエリではなくCTEを採用する。

**理由**
- 生成されたSQLが人間に読みやすい
- エラー箇所のトレースが容易

### CTE名とDiagAnchorの分離

CTE名とDiagAnchorは別物として管理する。

| | 役割 | 例 |
|---|---|---|
| CTE名 | SQL識別子。短く安定した形式 | `c1`, `c2`, ... |
| DiagAnchor | App NodeのID。エラー返却先 | `node-abc123` |

DiagはCTE名→DiagAnchorのマップで管理する：

```go
map[string]string // CTEName → DiagAnchor
// 例: {"c1": "node-xyz", "c2": "node-abc", "c3": "node-def"}
```

**生成例**

```
// IRの木
DeriveColumn(DiagAnchor="node-def", ColumnName="b", Expr=a*2)
  └── DeriveColumn(DiagAnchor="node-abc", ColumnName="a", Expr=x+1)
        └── TableScan(DiagAnchor="node-xyz", Table={Name:"users"})
```

```sql
WITH
c1 AS (
  SELECT * FROM users
),
c2 AS (
  SELECT *, x + 1 AS a
  FROM c1
),
c3 AS (
  SELECT *, a * 2 AS b
  FROM c2
)
SELECT * FROM c3
```

---

## compile時のschema伝播方針

compileは RelNode を再帰的に処理し、各ノードについて以下を行う。

1. Input を先に compile する
2. Input の output schema を受け取る
3. その schema を使って軽量ガードレールを実施する
4. 自ノードの output schema を計算する
5. SQL断片と output schema を次段へ渡す

Schema は compile の内部状態として扱い、Phase 2 では RelNode interface に schema API は持たせない。

### 列解決ルール

- Phase 2 では列名は case-sensitive な完全一致で解決する
- 同名列・スコープ・alias はサポートしない（単一テーブル前提）
- 列解決ロジックは compile の1か所に集約し、各所にベタ書きしない

---

## バリデーション責務

### App層でやるべきもの

- 型整合
- UI入力妥当性
- 関数引数の型チェック
- null / 未入力チェック
- 列選択UIの整合

### compile側でやる軽量ガードレール

App層バリデーションを信頼しつつ、lowerバグや将来の直接IR投入に備えて最低限のチェックを行う。

- `TableScan.Table.Name` が空でない
- `DeriveColumn.ColumnName` が空でない
- `DeriveColumn.ColumnName` が input schema の既存列と重複しない
- `Project.Columns` が空でない
- `ColumnRef.Name` が input schema に存在する
- `Project.Columns` の各列が input schema に存在する

---

## Phase 2 スコープと将来拡張方針

### Phase 2 のスコープ制限

- Phase 2 では単一 input の RelNode のみをサポートする
- JOIN や複数入力を持つ RelNode はサポートしない
- すべての ColumnRef は単一の input schema に対して解決される

### 将来の JOIN 拡張に対する設計方針

Phase 2 の実装がゴミにならないために、以下の境界だけ先に分離しておく。

- 列解決は compile の1か所に集約し、単一 input 前提をコード全体にベタ書きしない
- `compile内部表現（compiledRel）` に Schema を持たせ、将来 JOIN で情報が増えても struct 拡張で対応できるようにする
- `ColumnRef.Table` は Phase 2 では未使用だが、将来の source / alias 指定のために予約する
- `Column` は将来 `SourceID` / `SourceName` 等を追加できる構造にしておく
- `TableRef` は将来 Alias 等を追加できる struct のまま保つ

今やらないこと（Phase 3 以降）：

- `Join` ノードの実装
- 複数入力 schema の完全設計
- `table.column` 形式の列解決
- alias 必須化

---

## App層とIRの関係

### App Node → IR対応

| App Node | IR |
|---|---|
| TableComponent | TableScan |
| FilterComponent | Filter |
| ProjectComponent | Project |
| DeriveColumnComponent | DeriveColumn |
| ConstComponent | Literalとしてexprに埋め込まれる（独立ノードなし） |

---

## 実装順序

1. `expr/` 型定義
2. `rel/` 型定義
3. `sql/` コンパイラ（TableScan + Filter 最小構成）
4. DeriveColumn対応
5. Project対応
6. App層 lower 実装
