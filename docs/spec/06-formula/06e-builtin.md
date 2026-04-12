---
scope: docs/spec/06-formula/06e-builtin.md
status: confirmed
last_updated: 2026-04-12
summary: >
  Built-in Formulaの表示・読み取り専用制御の仕様。
  通常FormulaページをBuilt-in用にロックする方法を定義する。
key_decisions:
  - ツリー上のアイコンにBバッジを表示
  - ページ自体は表示するが全フィールド読み取り専用
  - ヘッダー下に固定バナー「Built-in Formula は編集できません」
  - 参照元バー・Testパネルは通常通り表示
depends_on:
  - docs/spec/06-formula/index.md
---

# 06e — Built-in Formula仕様

---

## Built-in Formulaの扱い

- ツリー上のアイコンに `B` バッジを表示
- ページは表示するが全フィールド読み取り専用（クリックしても何も起きない）
- ヘッダー下に固定バナー: `Built-in Formula は編集できません`
- 参照元バー・Testパネルは通常通り表示
