# formuflow UI設計 マスタードキュメント

> このドキュメントはPhase 4 UI仕様策定の神doc。
> 細部を詰めるたびにここに追記し、整合をここで取る。

---

## 1. ロール定義

### ITエンジニア
- DB接続管理（アイコンバーのDB接続パネルで設定）・DatabaseTableComponent設定
- SQL系ビルトインFormula（WHERE / SELECT / ChooseCols / Filter等）を直接使う
- FlowでSQL処理を隠蔽してドメイン向けモジュールを公開する
- デバッグ: SQL実行結果・API疎通・ColumnRefs/Conditionが意図通りか確認

### ドメインエンジニア（非IT）
- IT製のFormulaとFlowをライブラリから拾ってキャンバスに置く
- 汎用Formulaで純粋な計算ロジックを書く（SQLは見えない）
- デバッグ: エッジの値を虫眼鏡で追う・テストケース設定

> **境界線**: SQLを意識するか否か。Formulaエディタは共通だが、使えるビルトインのセットが層によって異なる。

---

## 2. 主要ユースケース

### シナリオA（IT）: `GetIVCharByTransistorId` を作る

1. **DB登録**: 左端アイコンバーのDB接続パネルで接続情報を設定（インフラ設定。Componentではない）
2. **DatabaseTableComponent定義**: DBとその中から持ってくるTableを定義。FlowにDBTableComponentを置くとedgeにDBTable本体と各 `column: ColumnRef` が生える
3. **ChooseCols**: `(table: DatabaseTable) -> ColumnRefs` でカラム選択
4. **Filter**: `(refs: ColumnRefs, cond: Condition) -> ColumnRefs` で `transistorId = {id}` の条件設定
5. **FlowでラップしてPublish**: DefaultInputComponent → ... → DefaultReturnComponent で閉じてドメイン向けに公開

### シナリオB（ドメイン）: トランジスタ特性の計算Flowを組む

1. IT製の `GetIVCharByTransistorId` をキャンバスに置く
2. Constで `transistorId` を入力・接続
3. 汎用Formulaで計算ロジックを組む（例: 増幅率の計算）
4. デバッグで各エッジの値を確認（虫眼鏡）
5. テストケースで期待値と照合

---

## 3. Componentカタログ

| Component | 説明 | 主な使用者 |
|---|---|---|
| `ConstComponent` | 単一スカラー定数。output 1本。ディレクトリで管理（例: `/phys_consts/R`） | 両方 |
| `ConstsComponent` | Key-Valueの塊。複数のスカラーをセットで定義。output複数本（例: `Vth`, `Vsat`）。数個想定 | 両方 |
| `DatabaseTableComponent` | DB接続を参照してTableを定義。edgeにDBTable本体 + 各`column: ColumnRef`が生える | IT |
| `FormulaComponent` | 式の定義。汎用 or SQL系ビルトインを含むもので分かれる | 両方 |
| `FlowComponent` | ComponentをFlowで組み合わせ。ネスト可・隠蔽・再利用 | 両方 |
| `DefaultInputComponent` | Flow実行の始点。`ConstsComponent`のAPI入力版（値の供給元がAPIかハードコードかの違いだけ）。空引数OK | 両方 |
| `DefaultReturnComponent` | Flowの出力終点。`DefaultInputComponent`と対称。`ConstsComponent`と同構造で値の受け取り先がAPIレスポンス | 両方 |
| `MapComponent` | FormulaをnD的に適用（全組み合わせ）。出力: nDMap | 両方 |
| `ZipComponent` | Formulaを要素ごとに適用。LengthMode指定。出力: Column固定 | 両方 |

> **DB接続はComponentではない**: インフラ設定として左端アイコンバーのDB接続パネルで一元管理。DatabaseTableComponentが「どのDB接続を使うか」を参照する形。

### Const系の関係
```
ConstComponent       単一スカラー。キャンバスに置いたらoutput 1本
ConstsComponent      Key-Valueの塊。本質はConstの集合
DefaultInputComponent ConstsComponentのAPI入力版。構造は同じ、値の供給元だけ違う
```

### Componentツリー構造

```
/builtin
  /math       — SUM, AVG, IF 等
  /sql        — ChooseCols, Filter, WHERE 等
/user
  /formulas
  /flows
  /dbtables
```

### Flow の閉じ方ルール
- FlowはDefaultInputComponent → ... → DefaultReturnComponentで**必ず閉じる**
- DefaultInputComponentは空引数OK（内部でConstを使う場合など）
- DefaultInput / DefaultReturn は`ConstsComponent`と同構造。供給元（API入力）か受け取り先（APIレスポンス）かの違いだけ

---

## 4. 画面構成

### レイアウト方針: VSCode型

```
┌──┬─────────────────────────────────────────┐
│  │  [tab1: flows/1] [tab2: formulas/3]  x  │
│左│─────────────────────────────────────────│
│端│                                         │
│ア│         メインエリア（タブで開く）          │
│イ│                                         │
│コ│                                         │
│ン│─────────────────────────────────────────│
│バ│         デバッグ・トレースパネル           │
│ー│                                         │
└──┴─────────────────────────────────────────┘
```

- **左端アイコンバー（固定）**: アイコンクリックでにゅっとパネルが出る（VSCodeのアクティビティバー）
- **メインエリア**: ブラウザライクなタブで複数ページを開く
- **デバッグパネル**: 下部に出る（実行時）

### 左端アイコンバーのパネル一覧

| アイコン | パネル内容 |
|---|---|
| ツリー | Componentツリー（全Component一覧・フォルダ構造） |
| 検索 | Component検索 |
| DB接続 | DB接続管理（接続一覧・死活確認・接続設定）。インフラ設定 |
| 設定 | アプリ設定 |

### メインエリア（タブで開くページ）

| ページ | URL | 内容 |
|---|---|---|
| Flowキャンバス | `/flows/:id` | ReactFlow。メイン作業場 |
| Formula編集 | `/formulas/:id` | 式エディタ（汎用 or SQL系で表示が変わる） |
| DatabaseTable設定 | `/dbtables/:id` | テーブル定義・参照するDB接続を選択 |
| DefaultInput設定 | `/inputs/:id` | [定義タブ] 引数名・型のペア定義 / [テストタブ] テストケース一覧・値の設定 |

### 遷移ルール
- Componentツリーからダブルクリック → 該当ページをタブで開く
- Flowキャンバス上でノードをダブルクリック → 該当Componentのページをタブで開く
- ツリーはファイルシステムライク（フォルダ構造でComponent管理）

### Componentの配置方針
- **全ComponentはDirツリーにフラットに存在する**。ネストは参照で表現し、実体は常にツリー上の1エントリ
- FlowComponentは内部で使うComponentへの参照を持つ。よってFlow専用のFormulaやConstsはフォルダで整理するケースが多い（例: `/flows/get_iv_char/input`, `/flows/get_iv_char/calc_gain`）
- どこからも参照されていない孤立Componentはツリー上で名前を灰色表示

### DefaultReturnComponentのedge
- input/outputの数は可変。型はedge接続先から推論
- わざわざReturn側で型宣言不要

---

## 5. Componentの共通フィールド設計

全てのComponentが持つフィールド：

```
Component
  id              — 内部識別子（DB保存・参照用）
  title           — 表示名。KaTeXのデフォルト表示名。未設定なら"untitled"
  katexSymbol     — optional。有効化したらtitleの代わりに数式中で使う記号
  katexFormat     — optional。式の書き方（\frac等）を指定

各port（input/output）
  varName         — edge上の変数名（識別子）
  value           — 実値（Const系はリテラル、DBRef等）
```

- `katexSymbol`はチェックボックスでenable/disable。デフォはtitleを使う
- titleを設定しないとKaTeX上も"untitled"になる。それはユーザーの責任

---

## 6. Formula Inspect

ネスト深すぎ問題の根本解決。FormulaとFlowのページ右端にトグルボタンがあり、クリックでInspectパネルが横にスライドして出る（GitHub Copilot Chatと同じ感覚）。編集画面と並列表示可能。

### 表示構造

```
[ComponentTitle]
──────────────────────────
  KaTeX式（現在の展開粒度で描画）

▼ gm  [FormulaComponent]
    KaTeX式（展開）
  ▶ Ids  [FormulaComponent]   ← 折りたたみ
  ▼ Vgs  [ConstsComponent]
    Vgs = 3.3
    Vth  [ConstComponent]     ← 参照なし、▼なし
```

- 参照を持つComponentは左に`▼/▶`で展開・折りたたみ
- 展開粒度に合わせてトップのKaTeXも段階的に展開される
- 末端（ConstComponent等）は値をそのまま表示

### URL / 表示方式
- URL: `/formulas/:id`, `/flows/:id`（既存ページのまま）
- Inspectはページ右端からスライドするサイドパネル（幅調整可能）
- デフォルトは閉じた状態

---

## 7. Formulaエディタの方針

Formulaエディタは共通UI。ビルトインはComponentツリーの `/builtin/math` `/builtin/sql` に整理されており、使う人だけ使う。モード切替・権限分離の概念は不要。

| 種別 | UIの形 |
|---|---|
| 汎用計算式 | テキストエディタ（Excel式ライク） |
| SQL系ビルトイン呼び出し | フォーム寄りのUI（FilterCondition等の専用入力） |

### Condition型の2種類

**FilterCondition** — WHERE句的。`column = value`形式。ITがDBから引くときに使う。テキスト入力で割り切り。DuckDBのWHERE句で書けることは全部できる。

**BranchCondition** — IF的。`A > B`, `IS NULL`等。ドメインも使う。ビルトイン（IF関数等）として提供しつつテキストでも書ける。

---

## 8. Map / Zip Component

### キャンバス上の表現
- **コンテナ型ノード**。他のノードより一回り大きく、内側にFormula/FlowComponentが収まる（Scratchのブロック的）
- Map / Zip どちらも見た目は同じ。挙動だけ違う

### 挙動の違い
| | MapComponent | ZipComponent |
|---|---|---|
| 適用方式 | 全組み合わせ（nD） | 要素ごと |
| 出力 | nDMap | Column固定 |
| 軸指定 | クリックでどのportをColumn軸にするか選択 | LengthMode指定（min/max/zero_pad/error） |

### 操作
- コンテナ外枠をクリック → どのportをColumn軸として使うか選択できるUI
- 選択されたportのedgeが `Column[T]` になる

---

## 9. デバッグ・トレース仕様

| 機能 | ドメイン | IT |
|---|---|---|
| エッジの値を虫眼鏡で確認 | ✅ | ✅ |
| テストケース設定（Input/Output） | ✅ | ✅ |
| SQL実行結果・クエリ確認 | ❌ | ✅ |
| API疎通確認 | ❌ | ✅ |
| ColumnRefs / FilterConditionの中身確認 | ❌ | ✅ |

---

## 10. 未決事項ログ

| # | 項目 | 優先度 |
|---|---|---|
| ~~1~~ | ~~Condition型のUI~~ | **クローズ**: FilterCondition（テキスト入力）/ BranchCondition（ビルトイン+テキスト）に2種類分離 |
| ~~2~~ | ~~DefaultReturnComponentの出力型設計~~ | **クローズ**: `ConstsComponent`と同構造。値の受け取り先がAPIレスポンスになるだけ |
| ~~3~~ | ~~表示API / iframeプレビュー仕様~~ | **クローズ**: formuflow本体スコープ外。必要なら呼び出し結果をHTMLにして返すラッパーエンドポイント1本で済む |
| ~~4~~ | ~~ITとドメインの権限分離~~ | **クローズ**: ビルトインはツリーに整理するだけで十分。モード概念不要 |
| ~~5~~ | ~~Map/ZipComponentのUI~~ | **クローズ**: コンテナ型ノード。外枠クリックでColumn軸選択 |
| ~~6~~ | ~~ColumnRefs型の入力UI~~ | **クローズ**: ColumnRefsは型。UI上は`Ref`バッジとして表示されるだけ。特別な入力UI不要 |

---

## 変更履歴

| 日付 | 内容 |
|---|---|
| 2026-04-02 | Phase 4初版作成。ロール・ユースケース・画面構成・Component一覧を確定 |
| 2026-04-02 | DatabaseComponentをカタログから削除。DB接続はインフラ設定としてアイコンバーパネルに移動。Componentツリー構造追加。未決#4クローズ |
| 2026-04-02 | ConstsComponent追加。Const系3種の関係を整理（Const / Consts / DefaultInput）。ノード見た目の方針確定（左端inputハンドル・右端outputハンドル・型バッジ文字表記） |
| 2026-04-02 | Component全フラット化確定。テストケースをDefaultInputのタブに統合。左アイコンバーからテスト削除。ReturnComponentのedge可変化 |
| 2026-04-02 | Componentの共通フィールド設計追加（id/title/katexSymbol/katexFormat）。Formula Inspect仕様追加（右サイドパネル・トグル式） |
| 2026-04-02 | Condition型を2種類に分離（FilterCondition/BranchCondition）。Map/Zipのコンテナ型ノード表現確定。未決#1・#5クローズ。ColumnRefs型UI未決#6として追加 |
| 2026-04-02 | 全未決クローズ。DefaultReturnをConstsと同構造に確定。ColumnRefsは型として整理。表示APIはスコープ外に。 |