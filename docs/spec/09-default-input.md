---
scope: docs/spec/09-default-input.md
status: draft
last_updated: 2026-04-13
summary: >
  DefaultInput設定ページ（/inputs/:id）の仕様。
  引数名・型のペア定義UIとTestパネル（右パネル、テストケース管理）を定義する。
  DefaultInputは複数Flowから参照可能。テストのinput値はDefaultInputに紐づき共有される。
  expected値・実行はFlowキャンバスのTestパネルで行う。
key_decisions:
  - 1つのFlowに配置できるDefaultInputComponentは1つのみ（ただしDefaultInput定義自体は複数Flowから参照可能）
  - ページ内タブなし。引数定義エリアのみのシンプル構成
  - テストケースのinput値はDefaultInputに紐づく（複数Flow間で共有）
  - テストケースのexpected値・実行はFlowキャンバスのTestパネルに紐づく（Flow別に管理）
  - DefaultInputのTestパネル: input値の管理のみ。expected値なし・実行系ボタンはグレーアウト・caseは常時展開
  - Flow側Testパネルでもcase追加可能（input+expected両方入力）。追加されたinput値はDefaultInput側に反映
  - expected未入力のcaseは実行不可（▶非活性・すべて実行でもスキップ）
  - Flowのテストパネルでのcase展開/折りたたみ: expected入力済み→展開、未入力→折りたたみ
  - caseカードはhover時に▶・📋・🚮/🔗をwrapper枠外右下に表示（hover判定はwrapper全体・折りたたみ時も有効）
  - 削除ボタンの挙動: 他Flow参照ありは🔗（リンクボタン）→ダイアログ確認→参照先expected一括削除→case削除
  - 🔗の参照判定: Flow側は「今開いているFlow以外」がexpectedを持つ場合、DefaultInput側は「すべての参照Flow」のうち1つでもexpectedを持つ場合
  - caseタイトルはダブルクリックで編集可。内部でindexを持ち、デフォルト名はcase {max_index + 1}
  - テストパネルのUIはFormulaのTestパネルと類似するが、データの所在・ライフサイクルが異なるため実装は独立
depends_on:
  - docs/spec/01-layout.md                 # 右パネル共通仕様
  - docs/spec/05-tabs-navigation.md        # タブ構成・ページ遷移
  - docs/spec/06-formula/06a-inputs.md     # 引数定義UIの共通構造
related_specs:
  - docs/spec/06-formula/06d-test-panel.md # TestパネルのUIは類似するが実装は独立
  - docs/spec/02-flow-canvas.md            # Flowキャンバス側のTestパネル仕様（追記予定）
open_issues:
  - 引数の型選択肢（Phase 5で確定）
  - 定義エリアの保存時警告ダイアログの詳細（Phase 5で06c-save.mdに準じて決定）
  - Flowキャンバス側のTestパネル詳細仕様（02-flow-canvas.mdに追記が必要）
  - テスト実行結果のpass/fail判定ロジック（期待値との比較方法、Phase 5）
  - Tbl等2D型のテストケース入力UI（Phase 5、田アイコン→モーダル方式で検討）
  - state diagramはmockup作成後に追記予定
---

# 09 — DefaultInput設定ページ仕様

対応モック: `docs/mockups/09-default-input/`（未作成）

---

## 概要

`/inputs/:id` で開くDefaultInputの設定ページ。

**DefaultInputの参照モデル:**
- 1つのFlowに配置できるDefaultInputComponentは1つのみ
- ただしDefaultInput定義自体は複数のFlowから参照可能（共有可能）

**テストケースのデータ所在:**
- `case.input値` → DefaultInputに紐づく（参照する全Flow間で共有）
- `case.expected値` → 各Flowに紐づく（Flow別に独立して管理）

ページ内タブはなし。引数定義エリアのみのシンプル構成。
テストケースの管理は `[Test]` ボタンで開くTestパネル（右パネル）で行う。

---

## ページ構造

```
┌─────────────────────────────────────────────┐
│  ▶  TransistorIVInput      /inputs/xxxx     │  ← ヘッダー
│     トランジスタIV特性計算の入力定義  [Test] │  ← 概要 + Testボタン
├─────────────────────────────────────────────┤
│  このDefaultInputは Flow "GetIVChar",        │  ← 参照元バー（参照あり時のみ）
│  "CalcTransistorGain" で使用されています     │
├─────────────────────────────────────────────┤
│  引数 (INPUTS)                              │
│  変数名          型                          │
│  transistorId   F64  [✕]                   │
│  vBias          F64  [✕]                   │
│  gain           Col  [✕]                   │
│  [+]                                        │
└─────────────────────────────────────────────┘
```

---

## ヘッダー

Formulaエディタ（06 index.md）と同構造。

- **タイトル**: クリックでインライン編集。確定: Enter / フォーカスアウト。キャンセル: Esc
- **パス表示**: 右上に `/inputs/:path` 形式（読み取り専用）
- **概要欄**: タイトル直下。1〜2行のテキスト入力
- **[Test]ボタン**: `?rightPanel=test` をトグル。右パネルとしてTestパネルをスライドイン

---

## 参照元バー

- このDefaultInputを参照しているFlowが存在する場合のみ表示
- 表示形式: `このDefaultInputは Flow "GetIVChar", "CalcTransistorGain" で使用されています`
- 各Flow名はリンク。クリックで該当FlowキャンバスをタブでOpen
- 複数Flow参照時は06f参照元バーと同スタイルで列挙

---

## 引数定義エリア

`docs/spec/06-formula/06a-inputs.md` と同構造。

| 要素 | 仕様 |
|---|---|
| 変数名 | 自由入力。重複不可 |
| 型 | デフォルト `F64`。`▼` プルダウンで変更（選択肢はPhase 5で確定） |
| 削除 | row hover時に変数名右端に `✕` ボタン表示 |
| 並び替え | D&Dで並び替え可。Flowキャンバスのポート順に反映 |
| 追加 | リスト下の `[+]` ボタン。追加時は変数名空・型 `F64` で行追加 |

変数名の重複チェックはリアルタイム。重複時は赤枠表示。

### 保存モデル

Formulaエディタ（06c-save.md）と同じdraft/publishedの2状態モデルに従う。
引数定義の変更をpublishすると、参照元全Flowのポートレイアウトとエッジへの影響が出る。
（保存時警告ダイアログ等の詳細はPhase 5仕様策定時に06c-save.mdに準じて決定）

---

## Testパネル（右パネル / DefaultInput側）

`[Test]` ボタンクリックで `?rightPanel=test` がトグル。右からスライド（01-layout.mdの右パネル仕様に準拠）。

UIはFormulaエディタのTestパネル（06d-test-panel.md）と**類似するが、データの所在・ライフサイクルが異なるため実装は独立**。

**このパネルで管理するもの: input値のみ。expected値はなし。**
実行系ボタンはすべてグレーアウト（実行・pass/fail判定はFlowキャンバス側のTestパネルで行う）。
caseカードは常時展開。

### テストケースカード

```
┌────────────────────────────────────┐
│ case 1                             │  ← PASS/FAILバッジなし（実行はFlow側）
│                                    │
│ transistorId  F64    [ 1.5      ]  │
│ vBias         F64    [ 2.0      ]  │
│ gain          Col    [ [1,2,3]  ]  │
└────────────────────────────────────┘
                          ▶ 📋 🚮/🔗  ← hover時のみ表示（枠外右下）
```

- **hover判定**: カード＋アイコン行を含むwrapper全体
- **caseタイトル**: ダブルクリックで編集可。デフォルト名は `case {max_index + 1}`
- **インライン編集**: 各フィールドをクリックで直接入力
  - `F64`: 数値入力
  - `Col[F64]`: `[0.1, 0.2, 0.3]` 形式のテキスト入力
- **対応型スコープ（Phase 4）**: `F64`・`Col[F64]`のみ。`Tbl`等の2D対応はPhase 5
- **PASS/FAILバッジ**: なし（input値のみ管理のため）
- **hover時に枠外右下に表示するアイコン**:
  - `▶` 単体実行（**グレーアウト**）
  - `📋` 複製（末尾に追加）
  - `🚮` 削除 / `🔗` リンクボタン（いずれかが表示される）

### 削除ボタンの挙動（DefaultInput側）

| 条件 | ボタン | 挙動 |
|---|---|---|
| すべての参照Flowがこのcaseのexpectedをもっていないとき | 🚮 | 確認なしで即削除 |
| 1つでも参照Flowがこのcaseのexpectedをもっているとき | 🔗 | ダイアログ表示 → OKで参照先expected一括削除 → case削除 |

**ダイアログ文言:**
`このcaseは "FlowA", "FlowB" で参照されています。参照解除しますか？（参照先のexpected値も削除されます）`

### テストケース追加

カード一覧の下に `+ テストケースを追加` ボタン。
追加時は引数定義から変数名・型を引いた空のケースを末尾に挿入。デフォルト名は `case {max_index + 1}`。

### すべて実行ボタン

パネル下部に `▶ すべて実行` ボタン。**グレーアウト**。

---

## Flowキャンバス側のTestパネル

`/flows/:id` キャンバス右上の `[Test]` ボタンで `?rightPanel=test`。

そのFlowのDefaultInputのinput値（共有）＋このFlow用のexpected値を合わせて表示・管理する。
case追加時はinput値・expected値の両方をこのパネルで入力可能。追加されたinput値はDefaultInput側にも反映される。

### テストケースカード（Flow側）

```
┌────────────────────────────────────┐  ← expected入力済み→展開表示
│ case 1                       PASS  │
│                                    │
│ transistorId  F64    [ 1.5      ]  │  ← input値（DefaultInputと共有・編集可）
│ vBias         F64    [ 2.0      ]  │
│ gain          Col    [ [1,2,3]  ]  │
│ ─────────────────────────────────  │
│ 期待値        F64    [ 0.6      ]  │  ← expected値（このFlowに紐づく・編集可）
└────────────────────────────────────┘
                          ▶ 📋 🚮/🔗

┌────────────────────────────────────┐  ← expected未入力→折りたたみ表示
│ case 2                        ▼   │
└────────────────────────────────────┘
                          ▶ 📋 🚮/🔗
```

- **展開/折りたたみ**: expected入力済み→デフォルト展開、expected未入力→デフォルト折りたたみ
- **折りたたみ状態でもhover判定（wrapper）は有効**（▶📋🚮/🔗が出る）
- **expected未入力caseの▶**: 非活性（実行不可）
- **すべて実行**: expected未入力caseはスキップ
- **実行エンドポイント**: `POST /flows/:id/run`
- 詳細仕様は `docs/spec/02-flow-canvas.md` に追記予定

### 削除ボタンの挙動（Flow側）

| 条件 | ボタン | 挙動 |
|---|---|---|
| 今開いているFlow以外のFlowがこのcaseのexpectedをもっていないとき | 🚮 | 確認なしで即削除（このFlowのexpected＋input値を削除） |
| 今開いているFlow以外の1つでもexpectedをもっているとき | 🔗 | ダイアログ表示 → OKで参照先expected一括削除 → case削除 |

---

## State Diagrams

（mockup作成後に追記予定）

---

## 未決事項

| # | 項目 |
|---|---|
| 1 | 引数の型選択肢（Phase 5で確定） |
| 2 | 定義エリアの保存時警告ダイアログの詳細（Phase 5で06c-save.mdに準じて決定） |
| 3 | Flowキャンバス側Testパネルの詳細仕様（02-flow-canvas.mdに追記） |
| 4 | テスト実行結果のpass/fail判定ロジック（期待値との比較方法、Phase 5） |
| 5 | Tbl等2D型のテストケース入力UI（Phase 5、田アイコン→モーダル方式で検討） |
