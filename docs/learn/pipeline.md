# pipeline（全体の流れ）

→ [全体俯瞰に戻る](./overview.md)

## 具体例で追う

ユーザーが「motorテーブルからmodel="A111"のrpmを取り出して、10倍した値を返す」フローを定義した場合。

---

### Step 1: App層がASTを作る

```
ASTNode{Type: formula, Value: "WHERE"}
  ├── ASTNode{Type: table_query, Value: "motor"}
  └── ASTNode{Type: operator, Value: "="}
        ├── ASTNode{Type: variable, Value: "model"}
        └── ASTNode{Type: literal,  Value: "A111", DataType: string}
```

→ [AST解説](./ast.md)

---

### Step 2: lowerがIRに変換する

```
rel.DeriveColumn{
    ColumnName: "__ff_result",
    Expr: expr.BinaryOp{Op: "*",
        Left:  expr.ColumnRef{Name: "rpm"},
        Right: expr.Literal{Value: 10, DataType: int},
    },
    Input: rel.Filter{
        Predicate: expr.BinaryOp{Op: "=",
            Left:  expr.ColumnRef{Name: "model"},
            Right: expr.Literal{Value: "A111", DataType: string},
        },
        Input: rel.TableScan{Table: {Name: "motor"}},
    },
}
```

→ [lower解説](./lower.md) / [IRノード解説](./ir-nodes.md)

---

### Step 3: sqlgenがSQLに変換する

```sql
WITH
c1 AS (
  SELECT * FROM motor
),
c2 AS (
  SELECT * FROM c1
  WHERE model = 'A111'
),
c3 AS (
  SELECT *, rpm * 10 AS __ff_result
  FROM c2
)
SELECT * FROM c3
```

1ノード = CTE 1個。木の葉（TableScan）から根に向かって順番にCTEが積まれる。

---

### Step 4: DuckDBが実行する

生成したSQL文字列をDuckDBに投げる。評価はすべてDuckDBが行う。
IRもsqlgenも「SQLを作るだけ」で計算はしない。

結果と一緒にDiag（エラー・警告）が返ってくる。
DiagAnchorを使ってApp NodeのIDに紐付けることで、UIがどのノードでエラーが起きたか表示できる。

---

## DiagAnchorの流れ

```
ASTNode.ID
    ↓ lower時にそのままコピー
rel.***.DiagAnchor
    ↓ sqlgen時にCTE名と紐付け
map[CTEName]DiagAnchor
    ↓ エラー発生時
Diagnostic{DiagAnchor: "node-abc123", Message: "...", Severity: "error"}
    ↓
UIが該当App Nodeをハイライト
```

---

## 2モードの違い

| | デバッグモード | APIモード |
|---|---|---|
| 実行単位 | Component単位で順次 | Flow全体を1クエリ |
| SQL本数 | Component数分 | 1本 |
| edge値 | キャプチャして返す | 返さない |
| 用途 | UIでの確認・デバッグ | 本番API呼び出し |

両モードともIR → SQL → DuckDBを通す。Diagを返すため。