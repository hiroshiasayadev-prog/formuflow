# formuflow 開発ロードマップ

> **コンセプト**: Formula・Table・Flowを組み合わせて計算フローを構築する。コードなしで非IT技術者でも扱えるデータ変換ツール。

---

## Phase 1 — 完了

funcexpr（Python実装）で代替済み。

---

## Phase 2: Tabular / Relational Layer 🚧 現在地

- Table入力
- Project / Filter / DeriveColumn
- Relational IR（最小版）
- バックエンド: Go
- データ層: DuckDB

### パッケージ構成

```
expr/    ← 式の木構造 + SQL断片生成（funcexprのGo再実装）
rel/     ← テーブル変換のIRノード定義
sql/     ← RelNode → SQL文字列のコンパイラ
```

### 実行フロー

```
App Node（DB保存） → lower → IR → compile → SQL → DuckDB実行
```

---

## Phase 3: SQL Compile

- DuckDB専用（方言対応は後回し）
- 対象オペレータ: Project / Filter / Join / Aggregate のみ

---

## Phase 4: 仕様策定

- UIの挙動・操作フローをmdで定義（What）
  - 例: エッジのI/O表示、型チェックの見せ方、デバッグのトレース表現
- 技術的実現可能性の疑問が出たら小さく検証してから仕様に落とす
- 反復あり。完了条件は「設計に進める確信が持てた時」

---

## Phase 5: 設計

- ドメインモデル定義（Component / Node）
- DB schema設計（PostgreSQL）
- API仕様（実行API、フロントI/F）

---

## Phase 6: 実装

- App層（バリデーション / lower / 実行API）
- UI（React + React Flow）
- ※ App層とUIをさらに分フェーズにするかはPhase 5完了後に判断

---

> **注記**: Phase 4以降は現時点で詳細が定義できないため粗い粒度でまとめている。Phase 4・5を進める中で、機能ごとにサブフェーズへ分割するか、進め方の方針を改めて定める必要がある。

---

## 保留・後回し

| 項目 | 備考 |
|------|------|
| funcexpr / funcexpr_xr のデモ | formuflow Phase 1と統合予定 |
| UI詳細仕様 | Phase 4完了後に詰める |
| SQL方言対応 | DuckDB完結が安定してから |
| PostgreSQL参照 | DuckDB ATTACH経由で後付け予定（ロジック変更最小） |
| テストケース設定機能 | — |
| KaTeXレンダリング | — |
| インテリセンス | — |