---
scope: docs/spec/07-formula-inspect.md
status: confirmed
last_updated: 2026-04-07
summary: >
  FlowページのInspectパネル（右パネル）仕様。
  キャンバス上ノードの👁クリックで開き、ツリー展開に連動してKaTeX式が
  段階的に展開される。末端まで展開することで式の全容を即座に把握できる。
key_decisions:
  - 対象は /flows/:id のみ（Formulaページには存在しない）
  - 右パネル（?rightPanel=inspect）、メインエリアを押し広げる方式
  - DefaultReturn selectboxは常時表示（1つでも表示、複数なら選択可）
  - KaTeXエリアはツリーの展開状態に1:1で連動する
  - ツリー各行: 左にcomponentアイコン（カラーテーマ準拠）、右にジャンプリンク↗
  - KaTeX: \textcolorでcomponentカラー色付け
  - FlowComponentもツリーで展開可能（中身のFlowに潜る）
depends_on:
  - docs/spec/01-layout.md            # 右パネル共通仕様（幅・スライド・URL制御）
  - docs/spec/03-component-nodes.md   # componentアイコン・カラーテーマ
related_specs:
  - docs/phase4/phase4-ui-design-master.md
open_issues:
  - Constsの展開表示（{title}({key}) → 値）の詳細確認（07c以降）
---

# 07 — Formula Inspectパネル仕様

対応モック: `docs/mockups/07-formula-inspect/`
- `07a-inspect-panel-open.html` — パネル開閉・スライドイン
- `07b-inspect-tree-katex.html` — Tree展開 + KaTeX連動（全4段階）
- `07c-inspect-multi-return.html` — DefaultReturn複数のselectbox切り替え（未作成）
- `07d-inspect-flow-nest.html` — FlowComponentネスト展開（未作成）

---

## 概要

Flowキャンバス上のノードに表示される👁ボタンをクリックすることで、
右パネルとしてInspectパネルが開く。

パネルはツリーUIとKaTeXエリアの2ペイン構成で、
ツリーを展開するにつれてKaTeX式が段階的に詳細化される。
これにより、ノードが多くなって見づらいFlowでも、
式の全容を即座かつ数式として把握できる。

---

## 表示コンテキスト

| 項目 | 値 |
|---|---|
| 対象ページ | `/flows/:id` のみ |
| URLパラメータ | `?rightPanel=inspect` |
| 発火イベント | キャンバス上ノードの👁ボタンクリック |
| パネル幅 | 280px（デフォルト）、左端ドラッグでリサイズ可 |
| 表示方式 | メインエリアを押し広げる（オーバーレイではない） |
| 閉じる | パネル右上の✕ボタン、または同じ👁ボタンを再クリック |

---

## パネル構成

```
┌─────────────────────────┐
│ INSPECT              ✕  │  ← ヘッダー
├─────────────────────────┤
│ OUTPUT                  │
│ [Id              ▾]     │  ← DefaultReturn selectbox（常時表示）
├─────────────────────────┤
│                         │
│   KaTeXエリア            │  ← ツリー状態に連動して更新
│                         │
├─────────────────────────┤
│ ▶ ⇢ transistor_iv_flow  │
│   ▶ fx calc_drain_cu... │  ← Treeエリア（展開可）
│   ...                   │
└─────────────────────────┘
```

---

## DefaultReturn selectbox

- DefaultReturnのinputポートを選択肢として列挙する
- 1つのみでも常にselectboxを表示する（条件分岐なし）
- 選択を切り替えるとツリーおよびKaTeXが切り替わる

---

## ツリー表示ルール

### ツリーのroot

表示中のFlowがrootになる。

### 展開可能なノード

| ノード種別 | 表示 | 展開したとき |
|---|---|---|
| FlowComponent | title | 中身のFlowに潜る（再帰） |
| FormulaComponent | title | formulaのinput portに繋がっているノードが子として出現 |
| Const | title | 値（スカラー）が子として出現 |
| Consts | `{title}({key})` | 値が子として出現 |

### 末端（展開不可）

| ノード種別 | 表示 |
|---|---|
| DefaultInputのport | port名（変数名） |
| Constの値 | 値そのもの（数値・文字列） |
| Constsの値 | 値そのもの |

### 各行の構成

```
[indent] [▶/▼ or 空白] [componentアイコン] [name]  | ↗
```

- **chevron**（▶/▼）: 展開可能なノードのみ表示、末端は空白
- **componentアイコン**: `03-component-nodes.md` のカラーテーマに準拠
- **`| ↗` アクションボタン**: hover時に出現。縦線区切りの右に↗ボタン、クリックでそのcomponentの編集ページへ遷移
- **インデント**: 階層ごとに16px増加

---

## KaTeX連動ルール

ツリーの展開状態とKaTeX表示は1:1で対応する。

| ツリー状態 | KaTeX表示 |
|---|---|
| `▶ flow` | `flow(DefaultInputのport名, ...)` |
| `▼ flow / ▶ formula` | `formula(各inputに繋がるノードのtitle, ...)` |
| `▼ flow / ▼ formula / ▶ Const` | formulaの式（Constはtitle） |
| `▼ flow / ▼ formula / ▼ Const` | formulaの式（Constは値） |

**表示ルール詳細:**

- 未展開ノード → そのノードの **title** で表現
- FormulaComponent展開 → そのFormulaの **式** （KaTeX）を使用
- Const展開 → title から **値** に置換
- DefaultInputのport → **port名**（末端、変数として表現）
- FlowComponent → 展開前はtitle、展開後は中身のFlowのKaTeXに置換

**カラーリング（`\textcolor` で色付け）:**

| 種別 | カラー |
|---|---|
| FlowComponent | `#5eead4` |
| FormulaComponent | `#a5b4fc` |
| Const / Consts | `#fcd34d` |
| DefaultInputのport名 | `#67e8f9` |

---

## 👁ボタンの仕様

- Flowキャンバス上の各ノードのheaderに表示
- クリックでInspectパネルを開き、そのノードを起点としたInspectを開始する
- アクティブ状態（パネル開中）はボタンをハイライト表示（Flowカラー `#5eead4`）
- 別ノードの👁をクリックすると切り替わる
- パネルを✕で閉じると👁のハイライトも解除される

---

## FlowComponentのネスト展開

FlowComponent（`⇢` アイコン）はツリー上で展開可能。
展開するとその参照先FlowのDefaultReturnを起点とした
サブツリーが展開される（再帰的に潜れる）。

KaTeXもそれに連動して、FlowComponentのtitleが
内部Flowの展開されたKaTeX式に置換される。
