# formuflow ドキュメント運用方針

> このdocはClaude（AIアシスタント）との協働における、ドキュメント管理の方針を定める。
> Claudeへの指示も兼ねているため、別会話のClaudeはこのdocを最初に読むこと。

---

## 1. 背景・目的

Phase 4以降、仕様docが増加し、1セッションで全文を渡すことが非現実的になった。
また、破壊的変更が複数docに波及するケースが増えてきた。

以下の2つの方針でこれを対処する：

1. **Front Matterによる索引化** — 各docの冒頭に要約ブロックを置き、Claudeが必要なdocだけを選択的に読める構造にする
2. **変更ログの集中管理** — 破壊的変更の波及範囲を`docs/changes/`に記録する

---

## 2. Front Matter仕様

### 形式

各docの**先頭**（`# タイトル`の直前）に以下のブロックを置く：

```markdown
---
scope: docs/spec/06-formula-editor.md
status: confirmed
last_updated: 2026-04-06
summary: >
  FormulaエディタのUI仕様。引数定義・式バリデーション・KaTeXプレビュー・
  Testパネル（右パネル）・Built-in読み取り専用・保存時警告ダイアログを定義する。
key_decisions:
  - 引数定義を先に行い、式バリデーションに利用する
  - KaTeXはオプトイン（デフォルトOFF）
  - Built-inはread-only（Bバッジ表示）
  - Testパネルは右パネル（?rightPanel=test）としてトグル
  - 引数変更時のpublishはダイアログ確認 → 影響Flowをツリーで赤字表示
depends_on:
  - docs/spec/01-layout.md            # 右パネル仕様
  - docs/spec/05-tabs-navigation.md   # タブ構成
  - docs/phase4/phase4-ui-design-master.md
related_specs:
  - docs/spec/09-default-input.md     # Testパネルの構造共通化
open_issues:
  - 出力型・引数型のプルダウン選択肢（Phase 5で確定）
  - 型不整合バリデーションロジック（Phase 5）
---
```

### フィールド定義

| フィールド | 必須 | 説明 |
|---|---|---|
| `scope` | ✅ | このdoc自身のパス |
| `status` | ✅ | `confirmed`（確定済み）/ `draft`（議論中）/ `wip`（作業中） |
| `last_updated` | ✅ | 最終更新日（YYYY-MM-DD） |
| `summary` | ✅ | 3〜5行。このdocが何を定義するかを端的に。Claudeがdoc選択に使う |
| `key_decisions` | ✅ | 重要な設計決定の箇条書き。変更時は必ず更新する |
| `depends_on` | 任意 | このdocが前提とするdocのパスとその理由（`#`コメントで） |
| `related_specs` | 任意 | 密接に関連するdoc（depends_onほど強い依存ではないもの） |
| `open_issues` | 任意 | 未決事項の簡易一覧（詳細は本文） |

### statusの基準

- `confirmed` — セッション内で合意が取れた。破壊的変更には変更ログが必要
- `draft` — 方向性は決まっているが細部が未確定
- `wip` — まだ議論中・大きく変わりうる

---

## 3. 変更ログ運用

### ファイル構成

```
docs/changes/
  BREAKING_CHANGES.md   ← 破壊的変更の一覧
```

`BREAKING_CHANGES.md`は追記専用（古いエントリは削除しない）。

### 書くタイミング

**変更の都度**、Claudeが書く。確定した変更をdocに反映するのと同じタイミング。

### エントリ形式

```markdown
## [YYYY-MM-DD] 変更タイトル（何が変わったか一言で）

**変更内容:**
変更の詳細。なぜ変わったかを含む。

**影響doc:**
- [x] docs/phase4/phase4-ui-design-master.md — 変更済み（セクション7更新）
- [x] docs/spec/06-formula-editor.md — 変更済み
- [ ] docs/spec/03-component-nodes.md — 要確認・未反映
- [ ] docs/spec/09-default-input.md — Testパネル共通化に影響する可能性
```

- `[x]` = 反映済み
- `[ ]` = 未反映（次回セッションで要対応）

### 「破壊的変更」の定義

以下のいずれかに該当する場合は変更ログを書く：

- 他docが依存しているキー概念（Component定義・URL設計・保存モデル等）の変更
- `confirmed`状態のdocを遡って修正する場合
- 複数docへの同時更新が必要な場合

---

## 4. Claudeへの作業指示

### 別会話でFront Matterを書く場合のプロンプト

以下をそのままコピーして別会話で使う：

---

```
あなたはformuflow（OSS視覚フロー計算ツール）プロジェクトのドキュメント管理を担当しています。
プロジェクトルートは C:\Users\imved\projects\formuflow です（filesystem MCPで参照可能）。

【タスク】
docs/doc-policy.md を読み、Front Matter仕様に従って、指定されたdocの冒頭にFront Matterブロックを追加してください。

対象doc: [ここにパスを書く。例: docs/spec/06-formula-editor.md]

【手順】
1. docs/doc-policy.md を読んでFront Matter仕様を把握する
2. 対象docを全文読む
3. Front Matterの各フィールドを埋める（summaryはdocの内容から正確に要約する）
4. doc先頭に挿入してファイルを上書きする
5. 「追加しました」と変更内容を簡潔に報告する

【注意】
- `# タイトル`の前に`---`で囲んだYAMLブロックを置くこと
- summaryは将来Claudeがdoc選択に使うため、このdocが「何を定義するか」を端的に書く
- key_decisionsは本文から重要な設計決定を拾う（議論の経緯ではなく「決定事項」）
- status判断: 未決事項がなければ`confirmed`、あれば`draft`
```

---

### セッション開始時のClaude向け手順

新しいセッションで作業を始める場合：

1. `docs/doc-policy.md`（このdoc）を読む
2. `docs/changes/BREAKING_CHANGES.md`を読み、`[ ]`（未反映）の項目を確認する
3. 必要なdocのFront Matter（`---`ブロック）を確認し、必要なものだけ全文取得する

### mermaidダイアグラムの注意事項

- ラベル内の改行は `\n` ではなく `<br/>` を使う（mermaidは `\n` を改行として解釈しない）

### md-sectionを使ったdoc読み込みパターン

`md-section` MCPツールを使うとmdファイルのセクションをピンポイントで取得できる。全文読みより大幅にtokenを節約できる。

**基本パターン:**

```
# 1. 見出し一覧を取得（構造把握）
md-section:list_headings
  path: C:\Users\imved\projects\formuflow\docs\spec\06-formula-editor.md

# 2. 必要なセクションだけ取得（部分一致でOK、#不要）
md-section:read_section
  path: C:\Users\imved\projects\formuflow\docs\spec\06-formula-editor.md
  heading: 保存
  include_subheadings: true  # デフォルトtrue。サブ見出しも含む
```

**推奨アクセス手順:**

1. `filesystem:read_text_file` + `head: 80` でFront Matter（`---`ブロック）だけ読む
2. `md-section:list_headings` で見出し一覧を把握する
3. `md-section:read_section` で必要なセクションだけ取得する
4. それでも足りない場合のみ全文取得する

**使いどころ:**
- 特定の決定事項だけ確認したい場合（例:「保存モデルの仕様だけ見たい」）
- 複数docをまたいで関連セクションだけ集めたい場合
- 変更の波及範囲を確認するために`depends_on`先のdocの一部だけ確認したい場合

### 変更ログを書くタイミング

Claudeは以下のケースで**自動的に**`BREAKING_CHANGES.md`にエントリを追記する：

- `confirmed`状態のdocを修正する場合
- 複数docに同時に変更を加える場合
- 他docの`depends_on`に含まれるdocを変更する場合

---

## 5. 現在のdoc一覧と状態

> このセクションはFront Matter追加作業の完了に合わせて更新する

| doc | status | Front Matter |
|---|---|---|
| `docs/architecture.md` | confirmed | ❌ 未追加 |
| `docs/roadmap.md` | confirmed | ❌ 未追加 |
| `docs/phase4/phase4-ui-design-master.md` | confirmed | ❌ 未追加 |
| `docs/spec/01-layout.md` | confirmed | ✅ 済み |
| `docs/spec/02-flow-canvas.md` | confirmed | ✅ 済み |
| `docs/spec/03-component-nodes.md` | confirmed | ✅ 済み |
| `docs/spec/04-sidebar.md` | confirmed | ✅ 済み |
| `docs/spec/05-tabs-navigation.md` | confirmed | ✅ 済み |
| `docs/spec/06-formula/index.md` | confirmed | ✅ 済み |
| `docs/spec/06-formula/06a-inputs.md` | confirmed | ✅ 済み |
| `docs/spec/06-formula/06b-expression.md` | confirmed | ✅ 済み |
| `docs/spec/06-formula/06c-save.md` | confirmed | ✅ 済み |
| `docs/spec/06-formula/06d-test-panel.md` | confirmed | ✅ 済み |
| `docs/spec/06-formula/06e-builtin.md` | confirmed | ✅ 済み |
| `docs/spec/07-formula-inspect.md` | confirmed | ✅ 済み |
| `docs/spec/08-dbtable-editor.md` | confirmed | ✅ 済み |
| `docs/spec/09-default-input.md` | draft | ✅ 済み |
| `docs/learn/overview.md` | confirmed | ❌ 未追加 |
| `docs/learn/expr.md` | confirmed | ❌ 未追加 |
| `docs/learn/ir-nodes.md` | confirmed | ❌ 未追加 |
| `docs/learn/lower.md` | confirmed | ❌ 未追加 |
| `docs/learn/pipeline.md` | confirmed | ❌ 未追加 |
| `docs/learn/ast.md` | confirmed | ❌ 未追加 |

Front Matter追加後は `❌ 未追加` → `✅ 済み` に更新する。

---

## 6. 命名規則（参考）

全サンプルコード・例示・ドキュメント内の変数名は**トランジスタ関連用語**を使う（例: `transistorId`, `transistor_iv_table`, `transistor_input`）。`motor`等の他ドメイン用語は使わない。
