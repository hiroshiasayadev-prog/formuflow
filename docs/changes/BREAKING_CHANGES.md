# formuflow 破壊的変更ログ

> 追記専用。古いエントリは削除しない。
> 形式は `docs/doc-policy.md` のセクション3を参照。

---

<!-- エントリはここから下に追記していく -->

## [2026-04-11] デバッグパネルのSQL/IRタブを常時表示に変更（Show Debug廃止）

**変更内容:**
`View > Show Debug` によるSQL/IRタブの表示切替機能を廃止し、PROBLEMS / SQL / IR の3タブを常時表示にした。
隠す複雑さに見合うメリットがなく、IT側も「エラー時のスクショを現場に送ってもらう」ユースケースでSQL/IRが見えた方が都合がよいため。
デフォルトアクティブタブはPROBLEMS。

**影響doc:**
- [x] docs/spec/01-layout.md — メニューバーのShow Debug記述削除、タブ構成から表示条件列を削除
- [x] docs/phase4/phase4-ui-design-master.md — セクション9のデバッグ仕様テーブル・タブ構成を更新、ドメイン/IT分割テーブルを機能一覧の箇条書きに置換
- [ ] docs/spec/10-debug-panel.md — 未作成（将来作成時に本変更を前提とすること）
