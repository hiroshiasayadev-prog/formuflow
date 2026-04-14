# formuflow 破壊的変更ログ

> 追記専用。古いエントリは削除しない。
> 形式は `docs/doc-policy.md` のセクション3を参照。

---

<!-- エントリはここから下に追記していく -->

## [2026-04-14] 12-error-feedback.md 新規作成・warnカラーをオレンジ→黄色に変更

**変更内容:**
Task 12のspec作業として `docs/spec/12-error-feedback.md` を新規作成。
Flowキャンバス上のエラー・型不整合フィードバックUI仕様（ノード枠赤/黄枠・エッジ赤/黄・Severityカラー定義）を定める。

あわせて、warnカラーを `#f97316`（オレンジ）から `#eab308`（黄）に変更。
理由: errorの赤（`#ef4444`）との視覚的対比を大きくし、warning（許容範囲）とerror（接続不可）を一見で区別できるようにするため。

10-debug-panel.mdのPROBLEMSタブインタラクションに、
逆方向ナビゲーション（ノード枠クリック→PROBLEMSタブ該当行フォーカス）を追記。

**影響doc:**
- [x] docs/spec/12-error-feedback.md — 新規作成
- [x] docs/spec/02-flow-canvas.md — 型バッジwarnカラー変更・参照先注記追加
- [x] docs/spec/10-debug-panel.md — 双方向ナビゲーション追記
- [x] docs/mockups/02-flow-canvas.html — warnカラー置換（`#f97316` → `#eab308`）
- [x] docs/doc-policy.md — doc一覧に12を追加

## [2026-04-13] 10-debug-panel.md 新規作成

**変更内容:**
Task 10のspec作業として `docs/spec/10-debug-panel.md` を新規作成。
BREAKING_CHANGESの未反映項目「docs/spec/10-debug-panel.md — 未作成（将来作成時に本変更を前提とすること）」を解消。
合わせてTask 10〜14のspec分割方針を決定した：
- `10-debug-panel.md`: 下部デバッグパネルUI（PROBLEMS/SQL/IR）← 今回作成
- `13-history-panel.md`: HISTORYパネルUI（右パネル・実行履歴カード・Export to test）
- `14-debug-execution.md`: デバッグ実行モデル（inputソース・edge値キャプチャ・canvas吹き出し）

主な決定事項:
- SQL/IRはFlowタブアクティブ時のみ有効。canvas現在状態JSONをbackendに投げてcompile（実行不要・draft可）
- タブ状態はfrontend localState管理（URLに乗せない）
- EXPLAIN不要（将来機能）
- IRはraw textベース（ツリーは将来機能）

**影響doc:**
- [x] docs/spec/10-debug-panel.md — 新規作成
- [x] docs/doc-policy.md — doc一覧に追加

## [2026-04-12] 08-dbtable-editor: state diagram追記・依存参照修正・sidebar validation設計方針決定

**変更内容:**
D-08-1〜D-08-4のstate diagram（およびflowchart）を08-dbtable-editor.mdに追記。
あわせて以下を変更:
- `depends_on`の`06-formula-editor.md`（旧単一ファイル・削除予定）を`06-formula/06c-save.md`に修正
- `last_updated`を2026-04-12に更新

D-08-3のpublishフローから「影響Flowツリー赤字表示」を意図的に除外。
理由: これはDBtable保存の責務ではなく、サイドバー（Componentツリー）が管理すべきValidation状態表示の話であるため。

**Sidebar Validation設計方針（今回決定・04-sidebar.mdに未反映）:**
- 表示: VSCode準拠（`✕ ComponentName` / `⚠ ComponentName`）
- 評価タイミング:
  - タブで開いているComponent → 常時評価
  - 変更が発生したComponent → 依存グラフを再帰的に辿り影響範囲を非同期評価（数loopごとにsleep挟み）
  - それ以外 → DBに保存済みのvalidationフラグをそのまま表示
- 分類:
  - `error`（✕）: 永続的・構造的（Formula引数変更・DBスキーマ変更によるエッジ無効など）
  - `warning`（⚠）: 一時的・環境依存（DB接続断など）
- 非同期評価の詳細はbackend設計時に詰める

**影響doc:**
- [x] docs/spec/08-dbtable-editor.md — state diagram追記・depends_on修正・last_updated更新
- [x] docs/spec/04-sidebar.md — Validation状態管理（表示・評価タイミング・分類）の仕様追記済み

## [2026-04-12] 07-formula-inspect: ツリー構造をcompo列挙からarg-compo対応の2段構造に変更

**変更内容:**
Inspectツリーの展開時の子表示を「展開可能ノードを単純に列挙」から「arg名行 + compo行の2段構造」に変更。
DAGにおいて同一compoが複数の引数に繋がるケースでargとcompoの対応が表現できなかったため。
あわせてKaTeX変数名の解決ルール（上流componentのtitle/port名の伝搬）、
未接続argの扱い（ツリー上はarg名行のみ、KaTeX上で暗め表示）、
FlowComponent展開時の内部FlowDefaultInputのport名利用、
D-07-1〜D-07-3のstatディアグラムを追記。

**影鿹doc:**
- [x] docs/spec/07-formula-inspect.md — ツリー構造変更・KaTeX連動ルール追記・state diagram追記・Front Matter更新
- [ ] docs/mockups/07-formula-inspect/07b-inspect-tree-katex.html — ツリーがarg-compo2段になっていないため要更新
- [ ] docs/mockups/07-formula-inspect/07d-inspect-flow-nest.html — 同上

## [2026-04-12] 06-formula-editor.mdをディレクトリ構成にリファクタ

**変更内容:**
`docs/spec/06-formula-editor.md`（単一ファイル）を`docs/spec/06-formula/`ディレクトリに分割した。
Formulaエディタは機能が多く（引数・式・KaTeX・保存・Testパネル・Built-in）、単一ファイルでは後から参照する仕様書（バックエンドAPI・DB設計等）が肥大化したdocを参照しにくくなるため。
各state diagram（D-06-1〜D-06-6）は対応する機能ファイルの末尾に埋め込み。

分割後の構成:
- `index.md` — 概要・ページ構造・ヘッダー・参照元バー・ショートカット・未決事項
- `06a-inputs.md` — 引数（INPUTS）・D-06-4
- `06b-expression.md` — 式エリア・出力型・KaTeX・D-06-5・D-06-6
- `06c-save.md` — 保存モデル・ダイアログ・D-06-1・D-06-2
- `06d-test-panel.md` — Testパネル・D-06-3
- `06e-builtin.md` — Built-in扱い

**影響doc:**
- [x] docs/spec/06-formula/（新規作成・全6ファイル）
- [x] docs/doc-policy.md — doc一覧を更新
- [ ] docs/spec/06-formula-editor.md — 旧ファイル削除（手動で行うこと）
- [x] docs/spec/07-formula-inspect.md — depends_onの参照先を06-formula/index.mdに更新（元々参照なしのため変更不要）
- [x] docs/phase4/phase4-ui-design-master.md — 06への参照があれば更新

## [2026-04-12] 02/03の責務分離：ノード仕様を03に一本化

**変更内容:**
02-flow-canvas.mdに混在していたノード単体UI仕様（共通構造・カラーテーマ・型バッジ・ポートレイアウト・FormulaノードKaTeX・選択状態）を03-component-nodes.mdに一本化した。
02はキャンバスレベルのインタラクション（エッジ・D&D・削除・Map/Zip）のみを担当する責務に整理。
あわせて03にD-03-1（ノード選択）・D-03-2（ハンドルhover）のstate diagramを追加した（後にD-02-5・D-02-6として02へ移動、インタラクション系は02に集約）。

採用基準:
- カラーテーマ: 03準拠（rgba値を持つ詳細な方）
- ポートレイアウト・FormulaKaTeX・選択状態: 02準拠（より新しい方）

**影響doc:**
- [x] docs/spec/02-flow-canvas.md — `## Componentノード 共通仕様` セクション削除、Front Matter更新
- [x] docs/spec/03-component-nodes.md — ノード仕様統合・state diagram追加、Front Matter更新
- [x] docs/phase4/phase4-ui-design-master.md — `## 3. Componentカタログ` の関連ファイルコメントを確認済み（spec/03-component-nodes.mdはすでに列挙されており追記不要と判断）

## [2026-04-11] デバッグパネルのSQL/IRタブを常時表示に変更（Show Debug廃止）

**変更内容:**
`View > Show Debug` によるSQL/IRタブの表示切替機能を廃止し、PROBLEMS / SQL / IR の3タブを常時表示にした。
隠す複雑さに見合うメリットがなく、IT側も「エラー時のスクショを現場に送ってもらう」ユースケースでSQL/IRが見えた方が都合がよいため。
デフォルトアクティブタブはPROBLEMS。

**影響doc:**
- [x] docs/spec/01-layout.md — メニューバーのShow Debug記述削除、タブ構成から表示条件列を削除
- [x] docs/phase4/phase4-ui-design-master.md — セクション9のデバッグ仕様テーブル・タブ構成を更新、ドメイン/IT分割テーブルを機能一覧の箇条書きに置換
- [ ] docs/spec/10-debug-panel.md — 未作成（将来作成時に本変更を前提とすること）
