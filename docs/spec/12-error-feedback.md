---
scope: docs/spec/12-error-feedback.md
status: confirmed
last_updated: 2026-04-14
summary: >
  Flowキャンバス上のエラー・型不整合フィードバックUI仕様。
  ノード枠の赤/黄枠表示、エッジの赤/黄表示、Severityカラー定義を定める。
  PROBLEMSタブとの双方向ナビゲーションは10-debug-panel.mdに集約。
  ドラッグ中のportフィードバック（✕ overlay）は02-flow-canvas.md D-02-2に集約。
key_decisions:
  - ノードエラーはバッジなし・枠色で表示（error=赤、warning=黄）
  - ノード枠クリック → PROBLEMSタブ該当行フォーカス（双方向は10に記述）
  - 接続後に型不整合が生じたエッジは赤エッジで残す（削除を促す）
  - warnカラーは黄色（#eab308）に統一（errorの赤との視覚的対比を明確にするため）
  - スコープはFlowキャンバスのみ（v0.1.0）
depends_on:
  - docs/spec/02-flow-canvas.md   # エッジ型バッジ・ハンドルD&Dフィードバック
  - docs/spec/10-debug-panel.md   # PROBLEMSタブ・双方向ナビゲーション
open_issues:
  - 型システムの本設計はPhase 5。現時点のerror/warn判定はUIフィードバック確認用ダミー
  - ノード枠のアニメーション（パルス等）は将来検討
---

# 12 — エラー・型不整合フィードバック仕様

対応モック: `docs/mockups/02-flow-canvas.html`（既存モックにwarnカラー修正を適用）

---

## Severityカラー定義

Flowキャンバス上のエラー表示で使うカラーを統一する。

| Severity | 用途 | カラー |
|---|---|---|
| error | 型error・compileエラー | `#ef4444`（赤） |
| warning | numeric implicit cast（型warn） | `#eab308`（黄） |

> warnをオレンジではなく黄色にする理由: errorの赤との視覚的対比を大きくし、
> warning（許容範囲）とerror（接続不可）を一見で区別できるようにするため。

---

## ノード枠エラー表示

### 表示条件

| 条件 | Severity | 枠色 |
|---|---|---|
| バックエンドcompileエラーに含まれるノード | error | 赤（`#ef4444`） |
| 接続後に型不整合エッジを持つノード | error | 赤（`#ef4444`） |
| warning（numeric cast）エッジのみを持つノード | warning | 黄（`#eab308`） |
| errorとwarningが混在する場合 | error優先 | 赤（`#ef4444`） |

### スタイル

```
box-shadow: 0 0 0 1.5px {severity-color}
```

通常のノード枠（`border: 1px solid rgba(255,255,255,0.08)`）の外側にglow風で重ねる。

### インタラクション

ノード枠クリック（エラー状態時）→ デバッグパネルをオープン（PROBLEMSタブ）し、
該当ノードに関連するエラー行にフォーカス。
双方向ナビゲーション仕様の詳細は `docs/spec/10-debug-panel.md` を参照。

---

## エッジエラー表示

### 接続後の型不整合エッジ

エッジ接続後（または接続先Formulaの引数型変更後）に型不整合が生じた場合、
エッジを切断せず赤エッジで残す。ユーザーに手動での修正を促す。

> ドラッグ中に型errorのポートへは接続できない（`✕` overlay）。
> 接続後に不整合が生じるケースは「Formula引数型変更」「DatabaseTableスキーマ変更」等。

### エッジのSeverity別スタイル

| Severity | エッジ線色 | 中点バッジ | バッジ表示内容 |
|---|---|---|---|
| ok | デフォルト（白系） | 緑（`#4ade80`） | 型名のみ（例: `Col`） |
| warning | 黄（`#eab308`） | 黄（`#eab308`） | cast形式（例: `F64→I32`） |
| error | 赤（`#ef4444`） | 赤（`#ef4444`） | cast形式（例: `Col→F64`） |

### 型チェックルール（ダミー・Phase 2時点）

```
同型         → ok
numeric同士  → warning（I32 / I64 / F32 / F64 間）
それ以外     → error
```

> 型システムの本設計はPhase 5で行う。

---

## state diagram

### D-12-1: ノード枠のエラー状態

```mermaid
stateDiagram-v2
    [*] --> normal

    normal  --> error   : compileエラー検出 / 接続後型error発生
    normal  --> warning : 接続後型warning発生（errorなし）
    error   --> warning : errorが解消・warningのみ残る
    error   --> normal  : 全エラー解消
    warning --> normal  : 全warning解消

    note right of error
        errorとwarningが混在する場合はerror優先
    end note
```

### D-12-2: エッジのエラー状態

```mermaid
stateDiagram-v2
    [*] --> ok

    ok      --> warning : 接続後にnumeric cast不整合が生じた
    ok      --> error   : 接続後に型error不整合が生じた
    warning --> ok      : 型が一致するよう修正された
    error   --> ok      : 型が一致するよう修正された
    error   --> warning : エラーがwarningレベルに緩和された

    note right of error
        エッジは切断しない。
        ユーザーが手動で修正するまで赤エッジで残る。
    end note
```
