# formuflow 破壊的変更ログ

> 追記専用。古いエントリは削除しない。
> 形式は `docs/doc-policy.md` のセクション3を参照。

---

<!-- エントリはここから下に追記していく -->

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
- [ ] docs/phase4/phase4-ui-design-master.md — `## 3. Componentカタログ` の関連ファイルコメントを要確認（実害はないが03が主体になった旨を明記してもよい）

## [2026-04-11] デバッグパネルのSQL/IRタブを常時表示に変更（Show Debug廃止）

**変更内容:**
`View > Show Debug` によるSQL/IRタブの表示切替機能を廃止し、PROBLEMS / SQL / IR の3タブを常時表示にした。
隠す複雑さに見合うメリットがなく、IT側も「エラー時のスクショを現場に送ってもらう」ユースケースでSQL/IRが見えた方が都合がよいため。
デフォルトアクティブタブはPROBLEMS。

**影響doc:**
- [x] docs/spec/01-layout.md — メニューバーのShow Debug記述削除、タブ構成から表示条件列を削除
- [x] docs/phase4/phase4-ui-design-master.md — セクション9のデバッグ仕様テーブル・タブ構成を更新、ドメイン/IT分割テーブルを機能一覧の箇条書きに置換
- [ ] docs/spec/10-debug-panel.md — 未作成（将来作成時に本変更を前提とすること）
