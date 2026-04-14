---
scope: docs/spec/10-debug-panel.md
status: confirmed
last_updated: 2026-04-13
summary: >
  下部デバッグパネルのUI仕様。PROBLEMS / SQL / IR の3タブ構成を定義する。
  SQL/IRはFlowのcanvas状態JSONをbackendに投げてcompileした結果を表示する。
  デバッグ実行（edge値キャプチャ）は14-debug-execution.mdで定義する。
key_decisions:
  - PROBLEMS / SQL / IR の3タブを常時表示（Show Debug機能は廃止済み）
  - デフォルト閉じ、エラー発生時に自動オープン（PROBLEMSタブ）
  - SQL/IRはFlowタブがアクティブな場合のみ有効（他ページでは「Flowを開いてください」空状態）
  - SQL/IRはタブ切り替え時にcanvas現在状態のJSONをbackendに投げてcompile・fetch（デバッグ実行不要）
  - draft状態のFlowでもSQL/IRは表示可能
  - EXPLAIN表示は不要（将来機能）、SQLはテキストそのまま表示
  - IRはraw textベース（ツリー表示は将来機能）
  - タブ状態はfrontend localState管理（URLに乗せない）
  - パネル×ボタンでSQL/IRのメモリclear
  - PROBLEMSの表示スコープは現在開いているタブのコンポーネントのみ
depends_on:
  - docs/spec/01-layout.md              # デバッグパネルの寸法・開閉仕様
  - docs/spec/02-flow-canvas.md         # Flowキャンバスとの連携（エラーフォーカス）
related_specs:
  - docs/spec/12-error-feedback.md      # エラー表示の詳細（ノード赤枠・エッジバッジ）
  - docs/spec/14-debug-execution.md     # デバッグ実行モデル（edge値キャプチャ・吹き出し）
open_issues:
  - SQL/IRのsyntax highlight（将来機能）
  - IRのツリー表示（将来機能）
  - EXPLAIN表示（将来機能）
  - FormulaページでのSQL/IR表示（v0.1.0では非対応・将来検討）
---

# 10 — デバッグパネル仕様

対応モック: `docs/mockups/10-debug-panel/`

---

## 概要

アプリ下部に固定されたデバッグパネル。PROBLEMS / SQL / IR の3タブで構成される。
IT向け情報（SQL・IR）とエラー一覧（PROBLEMS）を一箇所で確認できる。

寸法・開閉の基本仕様は `docs/spec/01-layout.md` に準拠する。

---

## パネル開閉

| 条件 | 挙動 |
|---|---|
| デフォルト | 閉じた状態 |
| 現在のタブでエラー発生 | 自動オープン（PROBLEMSタブ） |
| キャンバス右下の `Debug` ボタン | トグル開閉 |
| パネル右上の `×` ボタン | 閉じる + SQL/IR メモリclear |

- 高さ: デフォルト `200px`、上端ドラッグでリサイズ可（最小 `80px`、最大 `500px`）
- タブ状態（PROBLEMS / SQL / IR のどれがアクティブか）はfrontend localState管理

---

## タブ構成

### PROBLEMS タブ

現在開いているタブのコンポーネントに関するエラー・警告を一覧表示する。
アプリ全体のエラーは表示しない。

**フォーマット:**

```
✕ [Edge {ComponentName: EdgeName} — {ComponentName: EdgeName}]: type mismatch: Col → F64
⚠ [{ComponentName}]: implicit cast F64 → I32 may lose precision
```

- `✕` = error（赤）
- `⚠` = warning（黄）

**インタラクション（双方向ナビゲーション）:**
- 行クリック → 対象のComponentまたはEdgeをキャンバス画面中央にフォーカス + 赤枠表示
- ノード枠クリック（エラー/warning状態時）→ デバッグパネルをオープン（PROBLEMSタブ）し、該当ノードに関連するエラー行にフォーカス
- エラーが0件の場合: 「No problems detected.」を表示

---

### SQL タブ

現在開いているFlowのcanvas状態からcompileされたSQLを表示する。

**表示条件:**
- Flowタブがアクティブな場合: SQLを表示
- Flow以外のタブがアクティブな場合: 「Flow を開くと SQL が表示されます」空状態を表示

**fetch タイミング:**
- SQLタブに切り替えた瞬間、canvas現在状態のJSONをbackendに送信してcompile・fetch
- draft状態・未保存状態でも表示可能（保存済みのFlowレコードではなくcanvas現在状態を使う）
- 前回のfetch結果はメモリに保持し、canvas変更がない場合は再fetchしない（変更検知はcanvas stateのhash等で管理）
- パネル×ボタンで保持内容をclear

**表示内容:**
- 生成SQLテキスト（syntax highlightは将来機能）

---

### IR タブ

現在開いているFlowのcompile中間表現（IR）をraw textで表示する。

**表示条件・fetch タイミング:**
SQLタブと同様。

**表示内容:**
- IR dump テキスト（ツリー表示は将来機能）

---

## 空状態一覧

| 状態 | PROBLEMS | SQL | IR |
|---|---|---|---|
| Flow以外のタブがアクティブ | スコープ外のため非表示 | 「Flowを開くと表示されます」 | 「Flowを開くと表示されます」 |
| Flowタブだがエラーなし | 「No problems detected.」 | SQLテキスト | IR テキスト |
| compile失敗 | エラー一覧 | 「コンパイルエラーのため表示できません」 | 「コンパイルエラーのため表示できません」 |

---

## 未決事項

- SQL/IR のsyntax highlight（将来機能）
- IR のツリー表示（将来機能）
- EXPLAIN（DuckDB実行計画）表示（将来機能、手元でSQLコピペで代替可）
- FormulaページでのSQL/IR表示（v0.1.0では非対応。FormulaのSQLはFlowのSQLに統合されるため、Flowを開いて確認する運用とする。将来的には誘導メッセージ表示を検討）
- canvas変更検知の詳細実装（hash戦略等、Phase 5で確定）
