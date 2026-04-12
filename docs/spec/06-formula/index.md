---
scope: docs/spec/06-formula/index.md
status: confirmed
last_updated: 2026-04-12
summary: >
  Formula編集ページ（/formulas/:id）の概要・ページ構造・ヘッダー・参照元バー・
  ショートカット・未決事項を定義する。詳細仕様は06a〜06eを参照。
key_decisions:
  - /formulas/:id で開くFormulaの編集ページ
  - タブ構成はFORMULA（メイン編集）とTEST（単体テスト）の2つ
  - 他のFormula・Flowへの参照は不可（Formulaはプリミティブな計算のみ）
  - タイトル変更はリアルタイムで参照元FlowComponentに反映（内部参照はUUIDのため安全）
depends_on:
  - docs/spec/01-layout.md            # 右パネル仕様
  - docs/spec/05-tabs-navigation.md   # タブ構成・ページ遷移
related_specs:
  - docs/spec/06-formula/06a-inputs.md
  - docs/spec/06-formula/06b-expression.md
  - docs/spec/06-formula/06c-save.md
  - docs/spec/06-formula/06d-test-panel.md
  - docs/spec/06-formula/06e-builtin.md
open_issues:
  - 出力型・引数型のプルダウン選択肢（Phase 5で確定）
  - 型不整合バリデーションロジックの詳細（Phase 5）
---

# 06 — Formulaエディタ仕様

対応モック: `docs/mockups/06-formula-editor/`
- `06a-formula-editor-normal.html` — 通常状態
- `06b-formula-editor-katex-format.html` — KaTeXフォーマット有効
- `06c-formula-editor-error.html` — バリデーションエラー
- `06d-formula-editor-test-panel.html` — Testパネル展開
- `06e-formula-editor-builtin.html` — Built-in Formula
- `06f-formula-editor-referenced.html` — 参照元バーあり

---

## 概要

`/formulas/:id` で開くFormulaの編集ページ。タブ構成は `FORMULA`（メイン編集）と `TEST`（単体テスト）の2つ。

Formulaはプリミティブな計算のみを定義する。他のFormula・Flowへの参照は不可。複数Componentの組み合わせはFlowで行う。

式の中で使えるのは**引数（INPUTS）として定義した変数名**と**演算子・ビルトイン関数**のみ。他のFormulaやFlowを関数として呼び出すことはできない（バリデーションエラーとして検出される）。

---

## ページ構造

```
┌─────────────────────────────────────────────────────────┐
│  fx  CalcGainBandwidth         /formulas/calc_gain_...  │  ← ヘッダー
│      トランジスタの利得帯域幅を計算する          [Test]  │  ← 概要 + Testボタン
├─────────────────────────────────────────────────────────┤
│  このFormulaは FlowComponent "GetIVChar", "..." で       │  ← 参照元バー
│  使用されています                                         │  （参照あり時のみ）
├─────────────────────────────────────────────────────────┤
│  式 (FORMULA)                 │  出力型                  │
│  ┌─────────────────────────┐  │  [自動推論 ☑]  F64      │
│  │ ivChar + vBias / gain   │  │                         │
│  └─────────────────────────┘  │  KaTeX プレビュー        │
│  エラー表示エリア               │  ivChar + vBias / gain  │
│                               │                         │
│  引数 (INPUTS)                │  KaTeX フォーマット      │
│  変数名          型            │  [ カスタムフォーマット☐] │
│  ivChar         Col  [✕]     │                         │
│  vBias          F64  [✕]     │                         │
│  gain           F64  [✕]     │                         │
│  [+]                          │                         │
└───────────────────────────────┴─────────────────────────┘
```

---

## ヘッダー

- **タイトル**: クリックでインライン編集。確定: Enter / フォーカスアウト。キャンセル: Esc
- **パス表示**: 右上に `/formulas/:path` 形式で表示（読み取り専用）。UUIDは非表示
- **概要欄**: タイトル直下。1〜2行のテキスト入力。確定タイミングはタイトルと同じ

タイトル変更は参照元FlowComponentのノードタイトルにリアルタイム反映される（内部参照はUUIDのため、タイトル変更は安全）。

---

## 参照元バー

- タイトル・概要の下、タブバーの上に表示
- 参照元がある場合のみ表示（参照なしなら非表示）
- 表示形式: `このFormulaは FlowComponent "GetIVChar", "CalcTransistorGain" で使用されています`
- 各FlowComponent名はリンク。クリックで該当FlowのページをタブでOpen

---

## ショートカット

| キー | 動作 |
|---|---|
| `Ctrl-S` | publish（保存確定） |
| `Ctrl-Z` | undo |
| `Ctrl-Shift-Z` | redo |
| `Enter` | 入力確定（各フィールド） |
| `Esc` | 編集キャンセル（タイトル等） |

---

## 新規Formula作成時の初期状態

| フィールド | 初期値 |
|---|---|
| タイトル | `untitled` |
| 概要 | 空 |
| 式 | 空 |
| 引数 | 空（0行） |
| 出力型 | `None`（自動推論ON） |
| KaTeXフォーマット | OFF |

---

## 未決事項

| # | 項目 |
|---|---|
| 1 | 出力型・引数型のプルダウン選択肢（型システム確定後 Phase 5で決定） |
| 2 | 型不整合の具体的なバリデーションロジック（Phase 5） |
