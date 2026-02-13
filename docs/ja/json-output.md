# JSON 出力

自動化・CI/CD 連携用に、グローバルフラグ `--json` で JSON 形式の出力が可能です。

## 対応コマンド

- **`al diff --json`**: 追加・削除・アップグレード対象パッケージを JSON で出力
- **`al doctor --json`**: 診断結果と各チェックのステータスを JSON で出力
- **`al sync --plan --json`**: 実行予定のアクションとサマリーを JSON で出力

## 出力例（`al diff --json`）

```json
{
  "additions": [
    {"type": "addition", "provider": "brew", "name": "jq", "id": "formula:jq"}
  ],
  "removals": [],
  "upgrades": [],
  "has_drift": true
}
```

## 特徴

- JSON 出力の有無にかかわらず終了コードは一貫（`al diff` で差分検出時は 1、`al doctor` でエラー検出時は 1、正常時は 0）
- デフォルトの人間可読な出力は変更なし
- スクリプトやパイプライン処理に最適
