# 03 — Componentノード全種 仕様

対応モック: `docs/mockups/02-flow-canvas.html`（02と同一ファイルに全種収録）

---

## ノード共通構造

```
.node
  .node-header       ← アイコン + タイトル
  .node-katex        ← = {KaTeX式}（Formulaのみ）
  .node-body
    .ports-col.inputs
      .port-label    ← "input"
      .port-row × N  ← handle | type-badge | var-name
    .divider
    .ports-col.outputs
      .port-label    ← "output"
      .port-row × N  ← handle | type-badge | var-name
```

---

## カラーテーマ一覧

アクセントカラーはheader背景・アイコン背景・node-nameテキストのみに使用。

| Component | ヘッダー背景 | アイコン | name色 |
|---|---|---|---|
| Formula | `rgba(99,102,241,0.22)` | `#6366f1` | `#a5b4fc` |
| Flow | `rgba(20,184,166,0.18)` | `#14b8a6` | `#5eead4` |
| Const / Consts | `rgba(245,158,11,0.18)` | `#f59e0b` | `#fcd34d` |
| DatabaseTable | `rgba(244,63,94,0.18)` | `#f43f5e` | `#fda4af` |
| DefaultInput | `rgba(6,182,212,0.18)` | `#06b6d4` | `#67e8f9` |
| DefaultReturn | `rgba(168,85,247,0.18)` | `#a855f7` | `#d8b4fe` |
| Map / Zip | `rgba(129,140,248,0.18)` | `#818cf8` | `#c7d2fe` |

---

## 型バッジ

全Component・全エッジで共通。Component色とは完全に分離。

```
color:      #4ade80
background: rgba(74,222,128,0.15)
```

---

## Component別仕様

### ConstComponent
- アイコン: `C`
- input: なし
- output: `value`（スカラー型）

### ConstsComponent
- アイコン: `C`（Constと同テーマ）
- input: なし
- output: 複数（Key-Value各1本）

### DatabaseTableComponent
- アイコン: `DB`
- input: なし
- output: `table`（Tbl型）+ 各列（Ref型）

### FormulaComponent
- アイコン: `fx`
- headerとbodyの間にKaTeXエリア
  - 表示: `= {式}`（左辺省略、`data-katex`属性にLaTeX文字列）
- output var-name: Formulaのタイトルそのまま

### FlowComponent
- アイコン: `⇢`
- input/output: FlowのDefaultInput/Returnの定義に従う

### DefaultInputComponent
- アイコン: `▶`
- input: なし
- output: 引数定義に従う（可変）

### DefaultReturnComponent
- アイコン: `◀`
- input: 戻り値定義に従う（可変）
- output: なし

### Map / ZipComponent（Container）
- アイコン: `⊞`
- 構造: 通常のnode-bodyではなく`container-body`を使用
  - 左: inputポート列
  - 中央: `container-slot`（formulaをドロップするエリア）
  - 右: outputポート列
- 詳細仕様: `docs/spec/02-flow-canvas.md` の「Map/ZipコンテナノードUI」節を参照

---

## ハンドル

```
width: 9px
height: 9px
border-radius: 50%
border: 1.5px solid rgba(255,255,255,0.25)
background: #111118
```

- hover時: 拡大（`scale(1.5)`）、border色が緑（`#4ade80`）に
- ドラッグ中に型不整合と判定されたhandle: hover無効、`✕`をoverlay

---

## 未決・Phase 5以降

- 型表記: `Col[F64]` 形式の型パラメータ（型システム設計後に決定）
- ノードのリサイズ