# formuflow Phase 3 タスク

## 目的

Map / Zip を DuckDB SQL で実現するための前提条件を揃える。

- **Map**（全組み合わせ）→ CROSS JOIN で実現
- **Zip**（要素ごと適用）→ unnest + row_number で実現

---

## タスク一覧

### 3-1: Join IRノード追加（rel/）

`rel/` に `Join` ノードを追加する。

```go
type JoinKind string

const (
    JoinCross JoinKind = "CROSS"
    JoinInner JoinKind = "INNER"
    JoinLeft  JoinKind = "LEFT"
)

type Join struct {
    DiagAnchor string
    Kind       JoinKind
    Left       RelNode
    Right      RelNode
    On         Expr // CROSS の場合は nil
}
```

**備考**: MapはCROSS JOIN前提なので `On` はnil許容。将来のINNER/LEFTのために定義だけ先に入れる。

---

### 3-2: Aggregate IRノード追加（rel/）

`rel/` に `Aggregate` ノードを追加する。

```go
type AggFunc string

const (
    AggSum   AggFunc = "SUM"
    AggCount AggFunc = "COUNT"
    AggAvg   AggFunc = "AVG"
    AggMin   AggFunc = "MIN"
    AggMax   AggFunc = "MAX"
)

type AggExpr struct {
    Func  AggFunc
    Col   string // 集約対象列名
    Alias string // AS で付ける列名
}

type Aggregate struct {
    DiagAnchor string
    Input      RelNode
    GroupBy    []string  // 空の場合は全体集約
    Aggs       []AggExpr
}
```

**備考**: `GroupBy` が空のとき GROUP BY 句なし（全体を1行に集約）。

---

### 3-3: sqlgen Join対応

`compileJoin()` を `compile.go` に追加する。

生成SQL（CROSS）:
```sql
cN AS (
  SELECT * FROM c1, c2
)
```

生成SQL（INNER/LEFT）:
```sql
cN AS (
  SELECT * FROM c1
  LEFT JOIN c2 ON <predicate>
)
```

**ガードレール**:
- `Kind` が未知の値はエラー
- INNER/LEFT で `On` が nil はエラー
- Left / Right の出力スキーマで列名が衝突する場合はエラー（同名列は曖昧）

**output schema**: Left schema + Right schema の結合。列名重複時はコンパイルエラー。

---

### 3-4: sqlgen Aggregate対応

`compileAggregate()` を `compile.go` に追加する。

生成SQL:
```sql
cN AS (
  SELECT col1, col2, SUM(amount) AS total
  FROM cM
  GROUP BY col1, col2
)
```

**ガードレール**:
- `Aggs` が空はエラー
- `AggExpr.Col` が input schema に存在しない場合はエラー
- `GroupBy` の各列が input schema に存在しない場合はエラー

**output schema**: GroupBy列 + Aggs の Alias列。

---

### 3-5: lower Join対応

`ast.go` は変更なし。`NodeTypeFormula`のまま`node.Value`で判定する設計にした。
`lower.go` に `lowerCrossJoin()` / `lowerJoin()` を追加し、`lowerNode()` のswitchで振り分ける。

対応するビルトイン（ASTノードの `Value`）:

| Value | 変換先 | 備考 |
|---|---|---|
| `CROSS_JOIN` | `rel.Join{Kind: JoinCross}` | arg[0]: left table, arg[1]: right table |
| `JOIN` | `rel.Join{Kind: JoinInner}` | arg[0]: left, arg[1]: right, arg[2]: predicate |
| `LEFT_JOIN` | `rel.Join{Kind: JoinLeft}` | arg[0]: left, arg[1]: right, arg[2]: predicate |

**Mapのlower方針（暫定）**:

```
MapComponent(formula=f, axis=[col_a], scalar=[col_b])
  →
DeriveColumn(expr=f(col_a, col_b))
  └── Join(Kind=CROSS, Left=TableScan(a), Right=TableScan(b))
```

※ Mapの完全なlower設計はPhase 3完了後に別途詰める。

---

### 3-6: lower Aggregate対応

`lower.go` で `rel.Aggregate` に変換する。

対応するビルトイン:

| Value | 変換先 | 備考 |
|---|---|---|
| `GROUP_BY` | `rel.Aggregate` | 後述のAST形式で受け取る |

**ASTの形（確定）**:

```
GROUP_BY(table, "department", SUM(salary, "total"))
  - Children[0]: table_query
  - Children[1..]: string literal → GroupBy列名
               OR  formula（SUM/COUNT/AVG/MIN/MAX） → AggExpr
```

集約関数ノードの形:

```
ASTNode{Type: formula, Value: "SUM"}
  ├── ASTNode{Type: variable, Value: "salary"}  ← 集約対象列
  └── ASTNode{Type: literal,  Value: "total"}   ← alias
```

- `Children[1..]` を走査し、string literalはGroupByに、formulaはAggsに振り分ける
- Aggsが1つもない場合はlowerエラー
- SUM/COUNT/AVG/MIN/MAX以外の集約関数名はlowerエラー
- 集約関数の引数は `(variable, string literal)` の2つ固定

---

## 実装順序

```
3-1 → 3-2 → 3-3 → 3-4 → 3-5 → 3-6
```

IRノード定義（3-1/3-2）が先。sqlgen（3-3/3-4）はIRノードに依存。lower（3-5/3-6）はIRノードに依存するがsqlgenには依存しない。

---

## Phase 3完了後に解放されるもの

| Component | 実現方法 |
|---|---|
| `MapComponent` | CROSS JOINで全組み合わせ行を生成 → DeriveColumnで式適用 |
| `ZipComponent` | unnest + row_number で要素ごとに対応付け → DeriveColumnで式適用 |

**Zipの課題**: InputがColumnComponent（値を持つ）とQueryColumn（DB由来）が混在しうるため、unnestの使い方はPhase 3後に別途設計が必要。