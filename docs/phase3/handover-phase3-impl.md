# formuflow Phase 3 実装引き継ぎ

## 完了したこと

Phase 3（Join / Aggregate IRノード追加）が完了した。`go build`通過済み。

---

## この会話で追加・確定した内容

### rel/ に追加したノード

**Join**
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
    On         expr.Expr // CROSS の場合は nil
}
```

**Aggregate**
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
    Col   string
    Alias string
}
type Aggregate struct {
    DiagAnchor string
    Input      RelNode
    GroupBy    []string
    Aggs       []AggExpr
}
```

### sqlgen に追加した関数

- `compileJoin()`: CROSS / INNER / LEFT JOINのCTE生成。Left/Right列名衝突はエラー。
- `compileAggregate()`: GROUP BY + 集約関数のCTE生成。

### lower に追加した対応

`ast.go`は変更なし。`NodeTypeFormula`のまま`node.Value`で判定。

| ASTのValue | 変換先 |
|---|---|
| `CROSS_JOIN` | `rel.Join{Kind: JoinCross}` |
| `JOIN` | `rel.Join{Kind: JoinInner}` |
| `LEFT_JOIN` | `rel.Join{Kind: JoinLeft}` |
| `GROUP_BY` | `rel.Aggregate` |

**GROUP_BYのAST形式（確定）**:
```
ASTNode{Type: formula, Value: "GROUP_BY"}
  ├── ASTNode{Type: table_query, Value: <table>}
  ├── ASTNode{Type: literal, DataType: string, Value: <group_col>}  // 0個以上
  └── ASTNode{Type: formula, Value: "SUM" | "COUNT" | "AVG" | "MIN" | "MAX"}
        ├── ASTNode{Type: variable, Value: <col>}
        └── ASTNode{Type: literal, DataType: string, Value: <alias>}
```

### docs/learn/ を新設

バックエンドの各層を解説するドキュメントを追加した。

```
docs/learn/
  overview.md    ← 全体俯瞰・目次
  ast.md         ← ASTノード解説
  expr.md        ← exprノード解説
  ir-nodes.md    ← IRノード一覧（input/output/接続ルール）
  lower.md       ← lower変換ルール・具体例
  pipeline.md    ← ast→lower→IR→sqlgen→DuckDBの流れ
```

---

## 次にやること（Phase 4）

ロードマップ: `docs/roadmap.md`

Phase 4はUI仕様策定。UIの挙動・操作フローをmdで定義する。
技術的実現可能性の疑問が出たら小さく検証してから仕様に落とす。

**未解決の設計課題（Phase 4以降で詰める）**:

- MapComponent / ZipComponentのlower設計
  - Map: CROSS JOIN → DeriveColumnで式適用までの具体的なlower
  - Zip: unnest + row_number の設計（ColumnとQueryColumnの混在をどう扱うか）
- JOIN系ビルトイン（INNER / LEFT）をApp層でどう提供するか
  - ユーザー向けのUIとして何を見せるか未定

---

## 成果物ファイル

| ファイル | 内容 |
|---|---|
| `docs/phase3/phase3-tasks.md` | Phase 3タスク一覧（実装済み内容に更新済み） |
| `docs/learn/` | バックエンド各層の解説ドキュメント |

---

## 参考リンク

- Phase 2実装引き継ぎ: `docs/phase2/handover-phase2-impl.md`
- Phase 2仕様書: `docs/phase2/phase2-spec.md`
- Component設計・Map/Zip/型システム: https://claude.ai/chat/3b97d71d-1d0c-432b-8539-1fab42f0fd18