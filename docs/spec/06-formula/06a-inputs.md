---
scope: docs/spec/06-formula/06a-inputs.md
status: confirmed
last_updated: 2026-04-12
summary: >
  FormulaエディタのINPUTS（引数定義）仕様。変数名・型・追加・削除・並び替え・
  重複バリデーションを定義する。
key_decisions:
  - 変数名重複チェックはリアルタイム（赤枠表示）
  - 削除はhover時に✕ボタン表示
  - D&Dで並び替え可（port順に反映）
  - KaTeXに影響する文字（\, {, }, ^, _等）は変数名に使用不可
depends_on:
  - docs/spec/06-formula/index.md
---

# 06a — 引数（INPUTS）仕様

---

## 引数（INPUTS）

| 要素 | 仕様 |
|---|---|
| 変数名 | 自由入力。重複不可。KaTeXに影響する文字（`\`, `{`, `}`, `^`, `_`等）は不可 |
| 型 | デフォルト `F64`。`▼` プルダウンで変更。選択肢はPhase 5で確定（TBD） |
| 削除 | row hover時に変数名右端に `✕` ボタン表示 |
| 並び替え | D&Dで並び替え可。順番がFlowキャンバスのport順に反映される |
| 追加 | リスト下の `[+]` ボタン。追加時は変数名空・型 `F64` で行追加 |

変数名の重複チェックはリアルタイム。重複時はその行の変数名に赤枠表示。

---

## State Diagrams

### D-06-4: 引数（INPUTS）行の状態

```mermaid
stateDiagram-v2
    normal : normal
    hover : hover<br/>（✕ボタン表示）
    error : error<br/>（赤枠 — 変数名重複）

    normal --> hover : row mouseenter
    hover --> normal : row mouseleave
    hover --> normal : ✕クリック（行削除）
    normal --> error : 変数名が既存と重複
    error --> normal : 変数名を修正して重複解消
```
