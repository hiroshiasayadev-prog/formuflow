---
scope: docs/spec/06-formula/06d-test-panel.md
status: confirmed
last_updated: 2026-04-12
summary: >
  Formulaエディタの右パネルとして開くTestパネルの仕様。
  トグル動作・URLパラメータ対応・テスト実行を定義する。
key_decisions:
  - [Test]ボタンで?rightPanel=testをトグル
  - 01-layout.mdの右パネル仕様に準拠（右からスライド）
  - 引数の変数名・型はFormulaの定義から引く
  - 実行エンドポイント: POST /formulas/:id/run
depends_on:
  - docs/spec/01-layout.md                  # 右パネル共通仕様
  - docs/spec/06-formula/index.md
  - docs/spec/06-formula/06a-inputs.md      # 引数定義を参照
related_specs:
  - docs/spec/09-default-input.md           # Testパネルの構造共通化
---

# 06d — Testパネル仕様

---

## Testパネル（右パネル）

ヘッダー右端の `[Test]` ボタンクリックで `?rightPanel=test` がトグル。右からスライドして出る（01-layout.mdの右パネル仕様に準拠）。

DefaultInputの `[テストタブ]`（`docs/spec/09-default-input.md`）と同構造。引数の変数名・型をFormulaの定義から引く。

- テストケース一覧・値の設定・実行
- pass / fail 表示
- 実行エンドポイント: `POST /formulas/:id/run`（FlowのAPIと同形式）

---

## State Diagrams

### D-06-3: Testパネルのトグル

```mermaid
stateDiagram-v2
    closed : 閉じている<br/>（?rightPanel=なし）
    open : 開いている<br/>（?rightPanel=test）

    closed --> open : [Test]ボタンクリック
    open --> closed : ✕ボタンクリック
    open --> closed : [Test]ボタン再クリック
    closed --> open : URLに?rightPanel=testを付与して遷移
```
