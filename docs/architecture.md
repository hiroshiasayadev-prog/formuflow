# formuflow アーキテクチャ

## 全体構造

```mermaid
flowchart TD
    subgraph COMP["Component層（ユーザーが定義・操作する世界）"]
        subgraph PRIM["primitive"]
            Const["Const\nスカラー定数"]
            DBTable["DBTable\nDB参照"]
            Column["Column\n1D配列"]
            Row["Row\n1行データ"]
        end
        subgraph COMP2["composite"]
            Formula["Formula\n式の定義"]
            Flow["Flow\n組み合わせ・ネスト可"]
            Map["Map\nnD適用"]
            Zip["Zip\n要素ごと"]
        end
        FuncNote["FormulaがDB操作をする場合:\nFILTER / CHOOSECOLS / FIRST 等のビルトインFuncCallを使用\nDBTableと edge で繋いで初めてSQLとして解決される"]
    end

    subgraph EXEC["実行層（2モード）"]
        subgraph DEBUG["デバッグモード（UI）"]
            D1["Componentを1つずつ順番に実行"]
            D2["各edgeの入出力値をキャプチャ"]
            D3["虫眼鏡でフロントに返す（核心機能）"]
            D1 --> D2 --> D3
        end
        subgraph API["APIモード（本番）"]
            A1["Flow全体を1クエリにコンパイル"]
            A2["DuckDBで一括実行"]
            A3["結果返却"]
            A1 --> A2 --> A3
        end
    end

    subgraph COMPILE["コンパイル層（共通）"]
        Lower["lower\nApp Node → IR変換\nApp層でvalidate済み"]
        IR["IR\nrel + expr\nDiagAnchor付き"]
        SQL["SQL compile\nCTE生成\nschema伝播 + Diag"]
        Lower --> IR --> SQL
    end

    DuckDB[("DuckDB\nSQL実行（評価はここ）")]

    COMP --> EXEC
    EXEC --> COMPILE
    COMPILE --> DuckDB
    DuckDB --> DebugResult["デバッグ結果\nedge値 + Diag → UI"]
    DuckDB --> APIResult["API結果\n計算結果 → レスポンス"]
```

---

## データ型

```mermaid
classDiagram
    class Scalar {
        int
        float
        string
        bool
    }
    class Column {
        1D配列
    }
    class NDMap {
        2DMap
        3DMap
    }
    class QueryTypes {
        QueryRow
        QueryColumn
        QueryTable
        note: DB参照型（長さ・存在が不明）
    }
```

---

## Component詳細

```mermaid
flowchart LR
    subgraph primitive
        Const["Const\n値を自分で持つ"]
        DBTable["DBTable\nDBへの参照を持つ"]
        Column["Column\n値を自分で持つ"]
        Row["Row\n値を自分で持つ"]
    end

    subgraph composite
        Formula["Formula\n式定義\n引数: 任意の型"]
        Flow["Flow\nFormulaとDBTableをedgeで繋ぐ\nネスト可（隠蔽・再利用）"]
        Map["Map\nFormulaをnD的に適用\naxis引数 × scalar引数\n出力: nDMap"]
        Zip["Zip\nFormulaを要素ごとに適用\nLengthMode指定\n出力: Column固定"]
    end

    DBTable -->|"QueryRow\nQueryColumn\nQueryTable"| Formula
    Const -->|Scalar| Formula
    Column -->|Column| Formula
    Formula -->|Scalar結果| Flow
    Flow -->|Scalar / Column / nDMap| Map
    Flow -->|Column| Zip
```

---

## コンパイル層詳細

```mermaid
flowchart TD
    AppNode["App Node\n(DB保存済み)"]

    subgraph AppLayer["App層"]
        Val["バリデーション\n型整合 / null / 列存在 / 関数引数"]
    end

    subgraph IRLayer["IR (rel + expr)"]
        subgraph rel
            TableScan["TableScan\nDiagAnchor"]
            Filter["Filter\nDiagAnchor + Predicate:Expr"]
            Project["Project\nDiagAnchor + Columns"]
            DeriveColumn["DeriveColumn\nDiagAnchor + Expr"]
        end
        subgraph expr
            Literal["Literal\nValue: any + DataType"]
            ColumnRef["ColumnRef\nName + Table(予約)"]
            BinaryOp["BinaryOp\nOp: enum + Left + Right"]
            FuncCall["FuncCall\nName + Args"]
        end
    end

    subgraph SQLCompile["sql/ コンパイラ"]
        CTEGen["CTE生成\nc1, c2, c3..."]
        SchemaPropagate["schema伝播\n子→親の順に再帰"]
        DiagMap["DiagMap\nCTE名 → DiagAnchor"]
    end

    AppNode --> Val --> rel
    rel --> expr
    SQLCompile --> FinalSQL["WITH c1 AS (...)\nSELECT * FROM cN"]
    SchemaPropagate --> CTEGen
    DiagMap --> Diag["Diagnostic\n{ DiagAnchor, Message, Severity }"]
```

---

## 実行モード比較

| | デバッグモード | APIモード |
|---|---|---|
| 用途 | UIで確認・デバッグ | 本番APIとして呼び出し |
| 実行単位 | Component単位で順次実行 | Flow全体を1クエリ |
| edge値 | キャプチャして返す（虫眼鏡） | 返さない |
| パフォーマンス | 捨てる | 優先 |
| SQL本数 | Component数分 | 1本 |
