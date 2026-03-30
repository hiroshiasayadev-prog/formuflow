# formuflow Phase 2 実装引き継ぎ

## 次のチャットでやること

`expr/` → `rel/` → `sql/` の順でGoの型定義を書く。
仕様書が固まってるので基本的に写すだけ。

---

## 確定済み決定事項（この会話で追加）

### アーキテクチャ
- デバッグモード（UI）：Component単位で順次実行、各edgeの値をキャプチャしてJSONで返す
- APIモード（本番）：Flow全体を1クエリにコンパイルして実行
- **両モードともIR→SQL→DuckDBを通す**（有意なDiagを返すため）
- デバッグ結果の表示はUI側の責務。バックエンドはJSONで返すだけ

### Componentの全体像
**primitive/**
- ConstComponent：スカラー定数
- DatabaseTableComponent：DBテーブルへの参照
- ColumnComponent：1D配列
- RowComponent：1行データ

**composite/**
- FormulaComponent：式の定義。DBTableを引数に取れるがDB参照の解決はFlowの責務
- FlowComponent：FormulaとDatabaseTableをedgeで繋ぐ。ネスト可（隠蔽・再利用）
- MapComponent：FormulaをnD的に適用（全組み合わせ）。出力: nDMap
- ZipComponent：Formulaを要素ごとに適用。LengthMode指定。出力: Column固定

**データ型**
- Scalar: int / float / string / bool
- Column（1D配列）
- 2DMap / 3DMap
- QueryRow / QueryColumn / QueryTable（DB参照型、長さ・存在が不明）

### IRの目的
- Diagのためだけに存在する（ユーザーは触らない）
- FormulaのFuncCall（FILTER, CHOOSECOLS, FIRST等）がコンパイル先としてIRに落ちる

---

## 成果物ファイル

- `/mnt/user-data/outputs/phase2-spec.md`：IR・SQL設計の仕様書（型定義・設計方針すべて記載）
- `/mnt/user-data/outputs/architecture.md`：全体アーキのmermaid図

---

## phase2-spec.md の内容サマリ

### パッケージ構成
```
expr/  ←  rel/  ←  sql/
```

### expr/ 型定義（確定）
```go
type Expr interface { exprNode() }

type Literal struct {
    Value    any      // stringではなくany
    DataType DataType
}

type ColumnRef struct {
    Name  string
    Table string // optional、将来JOIN用に予約
}

type BinaryOp struct {
    Op    BinaryOpKind // enum化済み
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

### rel/ 型定義（確定）
```go
type Column struct {
    Name string
    Type DataType
    // future: SourceID, SourceName（JOIN対応時）
}
type Schema []Column

type SchemaProvider interface {
    GetSchema(tableName string) (Schema, error)
}

type RelNode interface { relNode() }

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
type Project struct {
    DiagAnchor string
    Input      RelNode
    Columns    []string // 既存列の選択のみ。rename/式評価不可
}
type DeriveColumn struct {
    DiagAnchor string
    Input      RelNode
    ColumnName string  // 既存列との重複はエラー
    Expr       Expr
}
```

### sql/ 型定義（確定）
```go
type CompileResult struct {
    SQL   string
    Diags []Diagnostic
}
type Diagnostic struct {
    DiagAnchor string
    Message    string
    Severity   string // "error" | "warning"
}

// compile内部表現
type compiledRel struct {
    CTEName    string  // c1, c2, c3...
    SQL        string  // このノードのCTE本体
    Schema     Schema
    DiagAnchor string
}

// compile関数シグネチャ（概念）
func compileRel(ctx *compileContext, node RelNode) (compiledRel, error)
func compileExpr(e Expr, schema Schema) (string, error)

type compileContext struct {
    provider SchemaProvider
    cteSeq   int
    ctes     []string
    diagMap  map[string]string // CTEName → DiagAnchor
}
```

### 設計方針（重要）
- CTE採用。CTE名（c1,c2...）とDiagAnchorは分離
- compile時にschema伝播：子compile→output schema取得→validate→親output schema計算
- RelNode interfaceにschema APIは持たせない（Phase 2）
- 列解決：case-sensitive完全一致、単一テーブル前提、compile内1か所に集約
- DeriveColumn.ColumnNameが既存列と重複したらエラー
- Literal/ColumnRefにDiagAnchorなし（App層バリデーションを信頼）
- JOIN未サポートだがColumnRef.Table予約・compiledRelにSchema持たせることで将来拡張を阻害しない

### compile側の軽量ガードレール
- TableScan.Table.Nameが空でない
- DeriveColumn.ColumnNameが空でない・既存列と重複しない
- Project.Columnsが空でない・各列がinput schemaに存在する
- ColumnRef.Nameがinput schemaに存在する

---

## 実装順序

1. `expr/` 型定義（Expr interface + Literal, ColumnRef, BinaryOp, FuncCall, DataType）
2. `rel/` 型定義（Schema, SchemaProvider, RelNode + 各ノード）
3. `sql/` コンパイラ（TableScan + Filter 最小構成）
4. DeriveColumn対応
5. Project対応
6. App層lower実装

---

## 参考リンク（過去会話）

- Excel関数・ビルトイン設計：https://claude.ai/chat/52c3c93b-115a-4849-b655-22790ae894a2
- Component設計・命名・Map/Zip/型システム：https://claude.ai/chat/3b97d71d-1d0c-432b-8539-1fab42f0fd18
