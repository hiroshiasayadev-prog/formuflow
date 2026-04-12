---
scope: docs/spec/06-formula/06c-save.md
status: confirmed
last_updated: 2026-04-12
summary: >
  Formulaの保存モデル（draft/published）・publish時の確認ダイアログ・
  引数変更時の影響Flow赤字表示を定義する。
key_decisions:
  - フォーカスアウト/Enterでdraft自動送信
  - Ctrl-Sでpublish確定
  - 引数の追加・削除・型変更があった場合はpublish時に確認ダイアログ
  - publish後、不整合が生じたFlowはツリー上でタイトルを赤字表示
depends_on:
  - docs/spec/06-formula/index.md
  - docs/spec/06-formula/06a-inputs.md   # 引数変更の検出
---

# 06c — 保存仕様

---

## 保存（draft / published）

- 式・引数・概要・タイトルの変更 → フォーカスアウト/Enterのタイミングでdraft自動送信
- Ctrl-S → publish確定
- **引数の追加・削除・型変更があった場合**: publish時に確認ダイアログを表示

```
保存の確認
以下のComponentに不整合が生じます:
  · GetIVChar
  · CalcTransistorGain
保存してよろしいですか？
[キャンセル]  [保存]
```

- 保存後、不整合が生じたFlowはツリー上でタイトルを赤字表示

---

## State Diagrams

### D-06-1: Formula全体の保存状態

```mermaid
stateDiagram-v2
    clean : clean
    draft : draft
    published : published

    clean --> draft : フォーカスアウト / Enter（自動送信）
    draft --> draft : フォーカスアウト / Enter（再送信）
    draft --> published : Ctrl-S（引数変更なし）
    draft --> published : Ctrl-S → 確認ダイアログ → 保存（引数変更あり）
```

### D-06-2: publish時のダイアログ分岐（引数変更あり）

```mermaid
flowchart TD
    A[Ctrl-S] --> B{引数変更あり?}
    B -- No --> C[そのままpublish]
    B -- Yes --> D[確認ダイアログ表示]
    D --> E{ユーザー選択}
    E -- キャンセル --> F[draft状態のまま]
    E -- 保存 --> G[publish確定]
    G --> H[影響Flowをツリーで赤字表示]
```
