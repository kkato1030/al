# ワークフロー例

## 環境の診断（セットアップ後や問題発生時）

```bash
al doctor
```

出力例:

```
[OK]    brew is available
[WARN]  link ~/.config/ghostty is broken
[WARN]  trial package review expired: gh (run 'al review')
```

## 試用から本番への昇格

1. `al add <パッケージ名>` で trial に追加
2. 評価後、`al promote <パッケージ名>` で昇格、不要なら `al remove <パッケージ名>`

## 仕事用・プライベートの分離

1. `al profile add work -e core.stable -p core.stable`
2. `al profile add private`
3. `al add <パッケージ> --prf work` / `al add <パッケージ> --prf private`
4. `al sync --profile work` または `al sync --all` で環境を揃える

## 変更内容を事前確認（プランモード）

1. `al sync --plan` でシステムへの変更を確認（実行なし）
2. 問題がなければ `al sync` で実際に適用

## GitHub での共有・復元

1. `al backup --init` でリポジトリ作成と push
2. 別マシンで `al sync owner/dotal` で clone して適用
3. 変更後は `al backup` で push
