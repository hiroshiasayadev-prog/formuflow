# formuflow UIモックタスク一覧

## 01 — アプリ全体レイアウト ✅
対象: VSCode型レイアウトの骨格
含むもの:
- 左端アイコンバー（固定）
- パネルエリア（にゅっと出る）
- メインエリア（タブバー＋コンテンツ）
- デバッグパネル（下部、折りたたみ）
対応mdドキュメント: `docs/spec/01-layout.md`

---

## 02 — Flowキャンバス ✅
対象: `/flows/:id` のメインキャンバス
含むもの:
- ReactFlowキャンバス背景（グリッド）
- 複数ノードが接続されたシーン（エッジ含む）
- エッジの型バッジ表示（接続線の中間に型を出すか否か）
- ノード選択状態（ハイライト）
- ズーム・パン操作エリア
対応mdドキュメント: `docs/spec/02-flow-canvas.md`

---

## 03 — Componentノード全種 ✅
対象: キャンバス上に置かれる各Componentの見た目
含むもの:
- FormulaComponent（indigo）
- FlowComponent（teal）
- ConstComponent（amber）
- ConstsComponent（amber・output複数）
- DatabaseTableComponent（rose）
- DefaultInputComponent
- DefaultReturnComponent
- Map/ZipComponent（コンテナ型）
対応mdドキュメント: `docs/spec/03-component-nodes.md`

---

## 04 — 左端アイコンバー＋パネル ✅
対象: サイドバーのアイコンバーと各パネル
含むもの:
- アイコンバー（ツリー・検索・DB接続・設定アイコン）
- Componentツリーパネル（フォルダ構造・孤立Component灰色表示）
- 検索パネル（Component検索）
- DB接続パネル（接続一覧・死活表示・設定）
対応mdドキュメント: `docs/spec/04-sidebar.md`

---

## 05 — タブバー＋ページ遷移 ✅
対象: メインエリアのタブと各ページのシェル
含むもの:
- タブバー（複数タブ・アクティブ・×ボタン）
- ツリーダブルクリック→タブで開く動線
- キャンバスノードダブルクリック→タブで開く動線
対応mdドキュメント: `docs/spec/05-tabs-navigation.md`

---

## 06 — Formulaエディタ ✅
対象: `/formulas/:id` の編集ページ
含むもの:
- 式入力エリア（テキストエディタ、Excel式ライク）
- 引数（input port）の宣言UI
- SQL系ビルトイン呼び出し時のフォームUI（FilterCondition入力）
- エラー・型不整合のインラインフィードバック
- KaTeXプレビュー
対応mdドキュメント: `docs/spec/06-formula-editor.md`

---

## 07 — Formula Inspectパネル ✅
対象: `/flows/:id` キャンバスのInspect右パネル
含むもの:
- 👁ボタンクリックでスライドイン（?rightPanel=inspect）
- DefaultReturn selectbox（output選択）
- KaTeXエリア（Tree展開状態に連動）
- Tree展開UI（Flow→Formula→Const の3段階、componentカラー）
- 各行のジャンプリンク（↗）
対応mdドキュメント: `docs/spec/07-formula-inspect.md`

---

## 08 — DatabaseTable設定ページ ✅
対象: `/dbtables/:id` の設定ページ
含むもの:
- 参照するDB接続の選択
- テーブル名入力・スキーマプレビュー（列一覧）
- 列ごとの型確認UI
対応mdドキュメント: `docs/spec/08-dbtable-editor.md`

---

## 09 — DefaultInput設定ページ ✅
対象: `/inputs/:id` の設定ページ（2タブ構成）
含むもの:
- [定義タブ] 引数名・型のペア定義UI
- [テストタブ] テストケース一覧・値の設定・実行
- テストケースのpass/fail表示
対応mdドキュメント: `docs/spec/09-default-input.md`

---

## 10 — デバッグパネル＋虫眼鏡 未着手
対象: 下部デバッグパネルとエッジ値確認UI
含むもの:
- デバッグパネル（展開・折りたたみ）
- エッジクリック時の虫眼鏡ポップオーバー（値表示）
- Scalar / Column / Table それぞれの値表示形式
- IT向け追加情報（SQL・ColumnRefs・FilterCondition）
対応mdドキュメント: `docs/spec/10-debug-panel.md`

---

## 11 — Map/ZipコンテナノードUI 未着手
対象: Map・ZipComponentのコンテナ型ノード
含むもの:
- コンテナ外枠（内側にFormulaノードが収まる）
- Column軸選択UI（外枠クリック時）
- LengthMode選択UI（Zip）
- nDMap / Column 出力ハンドルの表示
対応mdドキュメント: `docs/spec/11-map-zip-node.md`

---

## 12 — エラー・型不整合フィードバック 未着手
対象: 接続・実行時のエラー表示
含むもの:
- エッジ接続時の型不整合インジケータ
- ノード上のエラーバッジ
- デバッグパネルのDiagnosticリスト
- Severity（error/warning）の色分け
対応mdドキュメント: `docs/spec/12-error-feedback.md`
