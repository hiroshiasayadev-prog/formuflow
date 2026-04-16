---
scope: docs/ideas/python-service.md
status: draft
last_updated: 2026-04-17
summary: >
  Goでは実現困難な機能群（式変形・semantic search・AST構造類似検索）を
  担うオプショナルなPythonサービス基盤のアイデアメモ。
  Phase 7以降の候補。実装は未確定。
key_decisions:
  - Pythonサービスはオプショナル（起動していなければ該当機能はフォールバック）
  - プロセス間通信はHTTP（ランダムポート 49201〜、起動時にheartbeatでバックエンドに通知）
  - Goにこだわる理由はないためPythonサービスはFastAPI or Python直叩き
  - 式変形スコープはFormula onlyのflow限定（if/DB参照ノードは数学的スコープ外）
  - AST構造検索はSymPy正規化→全変数X置換→FastEmbed（軽量）でembedding化
  - 自然言語検索はsnowflake-arctic-embed-m-v2.0（多言語・高精度）を使用
  - ベクトルDBはQdrant in-process（Apache 2.0）。faiss不要
  - arctic-embed-m-v2.0はFastEmbedの.add_custom_model()でONNX経由ロード
open_issues:
  - 各機能の優先度・実装順序（未定）
  - Pythonサービスのライフサイクル管理（自動起動・終了）
  - arctic-embed-m-v2.0のONNXファイル存在確認（HuggingFace上）
  - Qdrantのindex永続化・更新タイミング
---

# Pythonサービス基盤（アイデアメモ）

> Phase 7以降の候補。ブレインストーミング段階。実装方針は仮。

---

## なぜPythonサービス基盤が必要か（Why）

Goエコシステムには数式処理・機械学習系のまともなライブラリがほぼ存在しない。
一方Pythonには SymPy（CAS）・FastEmbed（embedding）・Qdrant（ベクトル検索）など、
formuflowが必要とする機能群が揃っている。

これらをGoから呼び出すために、オプショナルなPythonサービスとして切り出す。

**設計方針:**
- Pythonサービスは**オプショナル**。起動していなければ該当機能は無効 or フォールバック
- 起動時に空きポート（49201〜）をランダムに選択し、heartbeatでGoバックエンドにポート番号を通知
- GoバックエンドはそのポートにHTTPで問い合わせる
- パフォーマンス要件は事実上ない（手動トリガー・低頻度）のでHTTPで十分

```
Pythonサービス起動
  → ポート 49201〜 をランダムに試行
  → 空きポート確定
  → heartbeat: POST /api/python-service/register {port: 49203}
  → Goバックエンドがポートを記憶
  → 以降 localhost:49203 に問い合わせ
```

---

## 何ができるか（What）

### 1. 式変形（Formula Solve）

Formulaの式を別の変数について解き、新しいFormulaを生成する。

**例:**
```
Formula: E = m * c**2
「cについて解く」を指定
→ 新Formula: c = sqrt(E / m)  （±の複数解をユーザーが選択）
```

**スコープ:**
- **対象**: Formula onlyで構成されたflow（数式として意味が成立するもの）
- **対象外**: if分岐・DBテーブル参照が含まれるflow
  - 理由: これらは「処理フロー」であり「式」ではない。「解く」という概念が数学的に存在しない

**flow自動再構成について:**
変形後の式をflow上のノードグラフに自動変換することは原理的には可能だが、
複雑な多項式ではノード構造が爆発する。
→ **変形結果は新しいFormulaコンポーネントとして登録するのみ**。flowへの組み込みはユーザーが手動で行う。

### 2. 自然言語によるFormula・Built-in検索

「電圧を計算するやつ」「正規化」などの自然言語クエリでComponentを検索する。

**使用モデル: `snowflake-arctic-embed-m-v2.0`**
- 多言語対応（日本語含む74言語）、英語性能も旧版より向上
- Apache 2.0ライセンス
- 113Mパラメータ、768次元
- FastEmbedの `.add_custom_model()` でONNX経由ロード（PyTorch不要）

```python
from fastembed import TextEmbedding
from fastembed.common.model_description import PoolingType, ModelSource

TextEmbedding.add_custom_model(
    model="Snowflake/snowflake-arctic-embed-m-v2.0",
    pooling=PoolingType.CLS,
    normalization=True,
    sources=ModelSource(hf="Snowflake/snowflake-arctic-embed-m-v2.0"),
    dim=768,
    model_file="onnx/model.onnx",
)
```

> ※ HuggingFace上のONNXファイル存在確認済

**ベクトルDB: Qdrant in-process**
- `QdrantClient(path="./qdrant_db")` でディスク永続化
- サーバー不要、Apache 2.0
- FastEmbedと直接統合できるため実装がシンプル

### 3. AST構造類似検索

「この式と構造が似ているFormulaを探す」。名前や変数名ではなく**式の形**で検索する。

**使用モデル: FastEmbed デフォルト（`BAAI/bge-small-en-v1.5` 等）**
- X置換済みの短い正規化文字列なので高精度モデル不要
- 軽量・高速優先

**実装方針（シンプルさ優先）:**

```
SymPyで正規化（展開・整理）
  → 全変数・定数を X に置換（ast.NodeVisitor で10行程度）
  → ast.unparse() で文字列化
  → FastEmbedでembedding化
  → Qdrantでベクトル検索
```

**例:**
```python
E = m * c**2  →  X = X * X**2
F = k * x**2  →  X = X * X**2   # 同じベクトル → 距離ほぼ0
```

構造が完全一致 → 距離ほぼ0、トップが同じで中身が複雑 → 距離近い、全然違う構造 → 距離遠い、という自然なランキングになる。

**先行研究との対応:**
code2vec（AST pathのembedding）が類似アプローチとして存在するが、
Javaコード向けの学習済みモデルしかなく数式には使えない。
数式特化の MathBERT もあるが「式+自然言語コンテキスト」込みの設計でformuflowの用途には過剰。
→ ルールベース正規化 + 軽量embeddingモデルの組み合わせが最もシンプルで実用的。

---

## どう実装するか（How）

### ライセンスまとめ

| ライブラリ | ライセンス | 備考 |
|------------|-----------|------|
| qdrant-client | Apache 2.0 | in-process DB含む |
| fastembed | Apache 2.0 | Qdrant製 |
| snowflake-arctic-embed-m-v2.0 | Apache 2.0 | 商用利用可 |
| sympy | BSD | |
| fastapi / uvicorn | MIT | |

全てOSSとして同梱・配布可能。

### Pythonサービス構成

```
python-service/
  main.py          # FastAPI エントリポイント
  solve.py         # SymPy 式変形
  search.py        # FastEmbed + Qdrant
  normalize.py     # AST正規化（X置換）
  requirements.txt # sympy, qdrant-client[fastembed], fastapi, uvicorn
```

### 各機能のI/O

**式変形:**
```json
POST /solve
{"expr": "E - m*c**2", "solve_for": "c", "variables": ["E", "m", "c"]}
→ {"solutions": ["-sqrt(E/m)", "sqrt(E/m)"]}
```

**AST構造検索:**
```json
POST /search/ast
{"expr": "a * b**2", "top_k": 5}
→ {"results": [{"formula_id": "...", "score": 0.98, "expr": "k * x**2"}, ...]}
```

**自然言語検索:**
```json
POST /search/semantic
{"query": "電圧を計算する", "top_k": 5}
→ {"results": [{"component_id": "...", "score": 0.87, "title": "..."}, ...]}
```

### ポート選択ロジック（Python側）

```python
import socket, random

def find_free_port(start=49201, end=49999):
    for _ in range(20):
        port = random.randint(start, end)
        with socket.socket() as s:
            if s.connect_ex(('localhost', port)) != 0:
                return port
    raise RuntimeError("空きポートが見つかりません")
```

---

## 優先度・実装時期

| 機能 | 優先度 | 備考 |
|------|--------|------|
| Pythonサービス基盤（heartbeat方式） | - | 他機能の前提 |
| 式変形 | 中 | 物理系ユーザーへの差別化機能 |
| AST構造類似検索 | 中 | 実装コスト低・ユーザー体験向上 |
| 自然言語検索 | 低 | 大量Componentが蓄積してから価値が出る |

いずれも **Phase 7以降**。Phase 5（ドメインモデル）・Phase 6（実装）が完了してから検討する。
