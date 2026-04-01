# formuflow バックエンド全体俯瞰

## 何を作っているか

ユーザーが定義した「式・フロー」を受け取り、SQLにコンパイルしてDuckDBで実行する仕組み。

```
ユーザーの式定義（App Node）
        ↓
    バリデーション（App層）
        ↓
      lower          ← ASTをIRに変換
        ↓
       IR            ← テーブル変換の木（rel + expr）
        ↓
     sqlgen          ← IRをSQL文字列に変換
        ↓
      DuckDB         ← SQL実行・評価
        ↓
     結果 + Diag
```

## 登場人物

| 名前 | 場所 | 一言 |
|---|---|---|
| AST | `domain/ir/ast/` | ユーザーの式をツリーで表現 |
| expr | `domain/ir/expr/` | スカラー計算の式ノード定義 |
| rel | `domain/ir/rel/` | テーブル変換のIRノード定義 |
| lower | `domain/ir/lower/` | AST → IR への変換処理 |
| sqlgen | `adapter/sqlgen/` | IR → SQL文字列への変換処理 |

## 各項目の詳細

- [AST](./ast.md) — ユーザーの式をどう表現するか
- [expr](./expr.md) — スカラー式のノード定義
- [IRノード](./ir-nodes.md) — テーブル変換のノード一覧（rel）
- [lower](./lower.md) — ASTをIRに変換する処理
- [pipeline](./pipeline.md) — 全体の流れを具体例で追う

## パッケージ依存関係

```
ast  ←  expr  ←  rel  ←  lower
                  rel  ←  sqlgen
                 expr  ←  sqlgen
```

- `ast` は `expr` の DataType を使う
- `rel` は `expr` の Expr を使う（例: Filter.Predicate が expr.Expr）
- `lower` は `ast` / `expr` / `rel` 全部を使う
- `sqlgen` は `rel` と `expr` を使う
- `domain/ir/expr` と `domain/ir/rel` はどのadapterにも依存しない