---
scope: docs/phase5/phase5-design-master.md
status: wip
last_updated: 2026-04-15
summary: >
  Phase 5の設計方針・タスクリスト・登場人物リスト・未決事項を管理するマスタードキュメント。
  spec ↔ architecture の整合検証を起点に、ドメインモデル・DB設計・API設計へ進む。
key_decisions:
  - 整合検証が Phase 5 最初の仕事（spec はバックエンド無視で作られているため）
  - 独自 YAML + render/validate スクリプト方式を採用（LLMコスト・汎用性の観点）
  - スキーマは既存 spec に当てはめながら育てる（先に詰めすぎない）
  - スクリプト言語は Go 推奨（未確定）
open_issues:
  - DefaultReturn / DefaultInput / Map・Zip の位置づけ（Component か修飾子か）
  - Project エンティティの有無
  - YAML スキーマの最小フィールド定義
  - スクリプト言語の確定
---

# Phase 5 設計マスター

---

## 1. 方針（Why）

### 1-1. なぜ整合検証から始めるか

Phase 4 の spec・mockup は UI 挙動の定義を優先して作られており、バックエンドの実現可能性を考慮していない。
この状態でドメインモデル・DB設計・API設計に入ると、すべてが「spec を後追いで正当化する作業」になり、矛盾が実装フェーズまで持ち越される。

そのため Phase 5 の最初の仕事は「現時点で分かる仕様を全て列挙し、論理的に破綻しない筋道を立てること」である。
破綻箇所は必ず出る。それを早期に特定・解消することがスタートライン。

### 1-2. 独自 YAML + スクリプト方式を採用する理由

**LLM に Mermaid DSL を直接書かせる方式の問題点：**

大規模な DAG・UML を LLM に自由記述させると、出力の一貫性が保てず、誤りの検出も困難になる。
複雑なドメインロジックを持つシステムの設計においても専用のバリデーション基盤が必要になるのは一般的な知見であり、
個人 OSS においてはなおさらコスト対効果が悪い。

**独自 YAML 方式の利点：**

- スキーマを決めることで LLM の出力自由度を絞り、機械検証が可能になる
- render スクリプトが Mermaid を生成するため、可視化コストは変わらない
- validate スクリプトで未定義参照・循環依存・矛盾を自動検出できる
- スキーマ自体が汎用資産になり、別プロダクトへの転用が可能
- 納期がない個人 OSS だからこそ、ここまで基盤を詰める価値がある

**想定スクリプト構成：**

```
scripts/
  validate.go   # 整合検証（未定義参照・循環依存・conflict 検出）
  render.go     # YAML → Mermaid md 生成
```

### 1-3. 「規格を先に詰めすぎない」方針

スキーマの粒度・単位は、実際の仕様の複雑さから逆算しないと決められない。
既存 spec に当てはめながら規格を育てる方が現実的であり、後から別プロダクトへ応用する際も実戦検証済みのスキーマの方が信頼できる。

ただし「これ以上決めてどうする？」というレベルまでは規格を詰めてからスタートする。
スタートラインに立てないまま作業を始めると必ず破綻する。

---

## 2. 登場人物（現時点・荒削り）

> Step A-2（概念整理）で更新する。曖昧なものには ⚠ フラグを立てる。

### 静的な「もの」

| 概念 | 説明 | 備考 |
|---|---|---|
| Component | Formula / Flow / DBTable / Const / Column / Row / Map / Zip の総称 | |
| Port | Component の入出力接続点（型を持つ） | |
| Arg | Formula の引数（名前・型・必須/任意） | Port との違いを要整理 |
| DataType | Scalar / Column / NDMap / QueryRow / QueryColumn / QueryTable | |
| BuiltinFunction | FILTER / CHOOSECOLS / FIRST 等 | |
| DBConnection | PostgreSQL 接続情報 | |
| Flow | Component を Edge で繋いだグラフ | |
| Edge | Port と Port を繋ぐ | |
| DefaultReturn | Flow の出力定義 | ⚠ Component か属性か未確定 |
| DefaultInput | Flow の入力口 | ⚠ Port との違い未確定 |
| Map | n次元適用 | ⚠ Component か Flow への修飾子か未確定 |
| Zip | 要素ごと適用 | ⚠ 同上 |
| Project | ユーザーが作る作業単位 | ⚠ エンティティとして存在するか未確定 |

### 実行時の「もの」

| 概念 | 説明 |
|---|---|
| IR | RelNode（TableScan / Filter / Project / DeriveColumn）+ Expr |
| SQL | CTE を連ねたクエリ |
| DiagAnchor | エラー位置の参照点 |
| Diagnostic | エラー・警告メッセージ（Severity 付き） |

### 永続化される「もの」

| 概念 | 説明 | 備考 |
|---|---|---|
| SaveState | draft / published | Component 単位で持つ |

### UI 固有の「もの」

| 概念 | 説明 |
|---|---|
| Tab | エディタのタブ（Component に対応） |
| RightPanel | ?rightPanel=xxx で制御されるパネル |
| InspectTree | Formula の inspect 表示（arg-compo 2段構造） |

---

## 3. タスクリスト

> 順番は依存関係を考慮している。並列可能なものは明記する。

### Phase 5-A：整合検証基盤の構築

- [ ] **A-1** 全 spec の Front Matter + 見出しをスキャンし、概念リストを作成する
- [ ] **A-2** 曖昧な概念（⚠ フラグ）を議論で潰し、登場人物リストを確定する（本ドキュメント §2 を更新）
- [ ] **A-3** YAML スキーマの最小フィールドを設計する（spec-atom / domain / db / api 各スキーマ）
- [ ] **A-4** `validate.go` と `render.go` の初版を実装する
- [ ] **A-5** 全 spec から仕様原子（spec-atom）を抽出して `docs/model/spec-atoms.yaml` に書き出す
- [ ] **A-6** validate を実行し、破綻箇所（未定義参照・循環・conflict）を一覧化する
- [ ] **A-7** 破綻箇所を一個ずつ議論・解消し、spec を修正する

### Phase 5-B：ドメインモデル定義（A-7 完了後）

- [ ] **B-1** `docs/model/domain-model.yaml` を設計する
- [ ] **B-2** Mermaid classDiagram を render して確認する
- [ ] **B-3** spec-atoms の `impl_requires` フィールドを埋めて整合確認する

### Phase 5-C：DB スキーマ設計（B-3 完了後）

- [ ] **C-1** `docs/model/db-schema.yaml` を設計する
- [ ] **C-2** Mermaid erDiagram を render して確認する

### Phase 5-D：API 仕様（B-3 完了後・C と並列可）

- [ ] **D-1** `docs/model/api-spec.yaml` を設計する
- [ ] **D-2** 実行 API・フロントエンド I/F を定義する

---

## 4. ファイル構成（予定）

```
docs/
  phase5/
    phase5-design-master.md   ← このドキュメント
  model/
    spec-atoms.yaml           # 全仕様原子（A-5 で作成）
    domain-model.yaml         # ドメインモデル（B-1 で作成）
    db-schema.yaml            # DB 設計（C-1 で作成）
    api-spec.yaml             # API 仕様（D-1 で作成）
scripts/
  validate.go                 # 整合検証（A-4 で作成）
  render.go                   # YAML → Mermaid 生成（A-4 で作成）
```

---

## 5. 未決事項

| # | 内容 | 優先度 | 対応タスク |
|---|---|---|---|
| U-1 | DefaultReturn の位置づけ（Component か Flow の属性か） | 高 | A-2 |
| U-2 | DefaultInput の位置づけ（Port との違い） | 高 | A-2 |
| U-3 | Map / Zip の位置づけ（Component か修飾子か） | 高 | A-2 |
| U-4 | Project エンティティの有無 | 中 | A-2 |
| U-5 | YAML スキーマの最小フィールド定義 | 高 | A-3 |
| U-6 | スクリプト言語（Go 推奨・未確定） | 低 | A-4 |
| U-7 | spec-atom の粒度（どこまで細かく切るか） | 高 | A-3 |
