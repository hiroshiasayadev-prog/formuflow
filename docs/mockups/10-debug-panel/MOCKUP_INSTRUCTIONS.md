# Task 10 — デバッグパネル mockup作成指示書

## 作業概要

`docs/spec/10-debug-panel.md` に基づき、デバッグパネルのインタラクティブHTMLモックを作成する。
6シナリオを **6つの別ファイル** に分けて作成すること。

---

## 事前準備（必ず最初に行うこと）

以下のファイルをMCP経由で読み込んでから作業を開始すること：

1. `docs/doc-policy.md` — Front Matter仕様・命名規則の確認
2. `docs/spec/10-debug-panel.md` — 今回のspec（必読）
3. `docs/spec/01-layout.md` — デバッグパネルの寸法・開閉仕様
4. `docs/mockups/08-dbtable-editor/08a-dbtable-normal.html` — デザインシステム・CSS変数・シェル構造の参照（スタイルはこれに合わせる）

---

## デザインシステム（08aから継承）

```css
:root {
  --bg:#111118; --bg2:#16161f; --bg3:#1c1c28; --bg4:#22222f;
  --border:rgba(255,255,255,0.08); --border2:rgba(255,255,255,0.14);
  --text:#e2e2e8; --text2:#8888a0; --text3:#55556a;
  --accent:#6366f1; --accent2:#a5b4fc;       /* indigo: Formula */
  --teal:#2dd4bf; --teal-bg:rgba(45,212,191,0.12);  /* teal: Flow */
  --green:#4ade80; --green-bg:rgba(74,222,128,0.15);
  --rose:#fb7185; --rose-bg:rgba(251,113,133,0.15);  /* rose: DatabaseTable */
  --amber:#fcd34d; --amber-bg:rgba(252,211,77,0.12); /* amber: Const */
}
body { font-family:'JetBrains Mono','Fira Code','Cascadia Code',monospace; font-size:12px; }
```

アプリシェルは以下の5エリア構成（VSCode型）：
- 左端アイコンバー（幅48px）
- サイドパネル（幅220px）
- メインエリア（タブバー + コンテンツ + デバッグパネル）
- 右パネル（今回は非表示でOK）

シェル全体の高さ: `660px` 程度。デバッグパネルはシェル下部に固定。

---

## シナリオボタンの配置

各ファイルの冒頭（`.shell` の外側）に `.scenario-row` としてシナリオ切替ボタンを配置する（08a等と同じパターン）。
シナリオボタンのアクティブスタイル：

```css
.scenario-btn.active { color:var(--accent2); border-color:rgba(99,102,241,0.5); background:rgba(99,102,241,0.1); }
```

---

## 作成ファイル一覧

### 10a — デバッグパネル閉じた状態
**ファイル名:** `10a-debug-closed.html`
**タイトル:** `10a — Debug Panel: 閉じた状態`

**シナリオ:**
- S1: 通常（Flowを開いている・エラーなし・デバッグパネル閉じ）← デフォルト

**表示内容:**
- アプリシェル全体（Flow `/flows/transistor-analysis` を開いている状態）
- キャンバス右下に `Debug ▲` ボタンを表示
- デバッグパネルは非表示（高さ0）
- `Debug ▲` ボタンクリックで `alert('→ 10b でパネル開状態を確認')` を表示

---

### 10b — PROBLEMSタブ・エラーあり（自動オープン）
**ファイル名:** `10b-problems-errors.html`
**タイトル:** `10b — Debug Panel: PROBLEMS（エラーあり）`

**シナリオ:**
- S1: エラー2件・警告1件 ← デフォルト
- S2: エラー1件のみ
- S3: 警告のみ2件

**表示内容:**
- デバッグパネルを開いた状態（高さ200px）、PROBLEMSタブがアクティブ
- タブバー: `PROBLEMS (3)` / `SQL` / `IR` / 右端に `×` ボタン
  - エラー+警告の合計件数をPROBLEMSタブラベルに `(N)` で表示
- PROBLEMSリスト（transistor系の用語を使うこと）:

```
✕  [Edge transistorAnalysis: vgs — ivCurve: vgs_input]: type mismatch: Scalar → Col
✕  [transistorGainCalc]: argument 'drain_current' is not connected
⚠  [ivCurve]: implicit cast F64 → I32 may lose precision
```

- `✕` は `var(--rose)` 色、`⚠` は `var(--amber)` 色
- 行hover時に背景を薄く変化させる
- 行クリック時: `alert('→ キャンバス上の対象コンポーネントにフォーカス')` で動作を示す
- パネル上端（リサイズハンドル）: hover時にカーソルを `row-resize` に変化させる
- `×` ボタンクリックでパネルを閉じる（高さ0にする）

**キャンバス上のノード表示（簡易で可）:**
- エラーのあるノード（`transistorGainCalc`）は外枠を `var(--rose)` 色にする

---

### 10c — PROBLEMSタブ・エラーなし
**ファイル名:** `10c-problems-clean.html`
**タイトル:** `10c — Debug Panel: PROBLEMS（エラーなし）`

**シナリオ:**
- S1: エラー0件・警告0件 ← デフォルト
- S2: エラー0件・警告1件

**表示内容:**
- デバッグパネル開いた状態、PROBLEMSタブがアクティブ
- タブバー: `PROBLEMS` / `SQL` / `IR` / `×`
  - エラー0件の場合はタブラベルにバッジなし
- S1: 「No problems detected.」をセンタリング表示（`var(--text3)` 色）
- S2: 警告1件のみリスト表示

---

### 10d — SQLタブ（Flow表示中）
**ファイル名:** `10d-sql-tab.html`
**タイトル:** `10d — Debug Panel: SQLタブ`

**シナリオ:**
- S1: SQL表示中（fetch済み）← デフォルト
- S2: fetch中（ローディング）
- S3: compile失敗

**表示内容:**
- デバッグパネル開いた状態、SQLタブがアクティブ
- タブバー: `PROBLEMS` / `SQL`（アクティブ）/ `IR` / `×`
- S1: 以下のようなSQLを表示（transistor系用語）:

```sql
-- transistor-analysis Flow
SELECT
  t.transistor_id,
  t.vgs,
  t.vds,
  (t.vgs - 2.0) * 0.8 AS drain_current,
  ((t.vgs - 2.0) * 0.8) * t.vds AS power_dissipation
FROM transistor_iv_table t
WHERE t.vgs >= 2.0
  AND t.vds BETWEEN 0.0 AND 10.0
```

- テキストは等幅フォント・行番号表示
- 右上に `Copy` ボタン（クリックでクリップボードコピー + ボタンラベルを `Copied!` に一瞬変化）
- S2: 「Compiling...」テキスト（スピナーがあればなおよい）
- S3: `var(--rose)` でエラーメッセージ「コンパイルエラーのため表示できません」+ 「PROBLEMSタブを確認してください」誘導テキスト

---

### 10e — IRタブ（Flow表示中）
**ファイル名:** `10e-ir-tab.html`
**タイトル:** `10e — Debug Panel: IRタブ`

**シナリオ:**
- S1: IR表示中（fetch済み）← デフォルト
- S2: fetch中

**表示内容:**
- デバッグパネル開いた状態、IRタブがアクティブ
- タブバー: `PROBLEMS` / `SQL` / `IR`（アクティブ）/ `×`
- S1: 以下のようなIR dumpをraw textで表示（transistor系用語）:

```
FlowNode {
  id: "transistor-analysis"
  nodes: [
    FormulaNode {
      id: "transistorGainCalc"
      inputs: [
        Arg { name: "vgs", type: F64, ref: "ivCurve:vgs" }
        Arg { name: "threshold", type: F64, ref: "const:vth" }
        Arg { name: "gain_factor", type: F64, ref: "const:k" }
      ]
      expr: "(vgs - threshold) * gain_factor"
      output_type: F64
    }
    DatabaseTableNode {
      id: "ivCurve"
      table: "transistor_iv_table"
      connection: "lab-db"
      columns: [vgs: F64, vds: F64, id: F64]
    }
  ]
  edges: [
    Edge { from: "ivCurve:vgs", to: "transistorGainCalc:vgs", type: Col(F64) }
    Edge { from: "const:vth", to: "transistorGainCalc:threshold", type: Scalar(F64) }
  ]
}
```

- テキストは等幅フォント・raw text（ツリー表示は将来機能）
- 右上に `Copy` ボタン（10dと同様）
- S2: 「Compiling...」テキスト

---

### 10f — Flow以外のページを開いている場合の空状態
**ファイル名:** `10f-non-flow-empty.html`
**タイトル:** `10f — Debug Panel: Flow以外（空状態）`

**シナリオ:**
- S1: Formulaページを開いている ← デフォルト
- S2: DBTableページを開いている

**表示内容:**
- デバッグパネル開いた状態
- PROBLEMSタブ: 現在タブのコンポーネント単体のエラーを表示（「No problems detected.」で可）
- SQLタブ: 「Flow を開くと SQL が表示されます」（`var(--text3)` 色、センタリング）
- IRタブ: 「Flow を開くと IR が表示されます」（同上）
- タブ切り替えボタンで PROBLEMS / SQL / IR の3状態を確認できること

---

## 共通実装ルール

1. **命名規則**: サンプルの変数名・テーブル名・Component名はすべてトランジスタ系用語（`transistorId`, `transistor_iv_table`, `ivCurve`, `transistorGainCalc` 等）を使うこと。`motor` 系は一切使わない。

2. **シェル構造**: 08a等の既存mockupと同じCSS変数・シェル構造を継承する。アイコンバー・サイドパネル・タブバーは簡易でよいが、全体のレイアウト感が伝わるように入れること。

3. **デバッグパネルの構造**:
```
.debug-panel
  .debug-resize-handle   ← 上端、height:4px、cursor:row-resize、hover時に薄くハイライト
  .debug-tabbar          ← PROBLEMS / SQL / IR タブ + 右端に × ボタン
  .debug-content         ← タブコンテンツ（overflow-y:auto）
```

4. **パネル高さ**: デフォルト200px。リサイズハンドルはhover時のカーソル変化のみ示せばOK（実際のドラッグ動作の実装は任意）。

5. **各ファイルは独立して動作すること**（外部CSSファイル参照なし・すべてインライン）。

6. **ファイル保存先**: `docs/mockups/10-debug-panel/` 配下に保存すること。

7. **mockup-tasks.mdの更新**: 全ファイル作成後、`mockup-tasks.md` のTask 10の末尾を `未着手` → `✅` に変更すること。

---

## 作業完了後の確認チェックリスト

- [ ] 6ファイルすべて作成済み（10a〜10f）
- [ ] motor系用語が一切含まれていないこと
- [ ] 各ファイルがブラウザ単体で開いて動作すること
- [ ] シナリオボタンで状態切替が機能すること
- [ ] `docs/mockups/mockup-tasks.md` のTask 10を ✅ に更新済み
