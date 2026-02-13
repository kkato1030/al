# Brewfile からの移行

既存の Brewfile を al に取り込む。デフォルトは登録のみ（既存環境を al に寄せる想定）。

## 事前準備

- 登録先 profile を用意: `al profile add <profile>`。Brewfile の種類に応じて `al provider add brew` / `al provider add mas` を実行。

## Brewfile から取り込む場合

```bash
al import Brewfile --prf core --dry-run   # 確認
al import Brewfile --prf core             # 登録のみ
al import Brewfile --prf core --install   # 未インストール分もインストール
```

## brew/mas から自動検出する場合（Brewfile 指定なし）

```bash
al import --prf core --dry-run   # 現在インストール済みパッケージを確認
al import --prf core            # 自動検出して登録
al import --prf core -i          # 対話的に選択してインポート
```

**注意**: 自動検出では、明示的にインストールしたパッケージのみを取り込みます（`brew leaves` を使用）。依存関係として自動インストールされたパッケージは含まれません。Cask と mas はすべて明示的インストール扱いです。

## オプション

| オプション | 説明 |
|------------|------|
| `--profile`, `--prf` | 登録先 profile（必須） |
| `-s`, `--stage` | stage 名 |
| `--dry-run` | 書き込まずパース結果と登録予定の一覧のみ表示 |
| `--install` | 未インストールのパッケージを brew/mas でインストール（Brewfile 使用時） |
| `--overwrite` | 同一 id・provider・profile の既存登録を上書き |
| `--verbose` | 対応外の行をスキップした理由を表示（Brewfile 使用時） |
| `-i`, `--interactive` | 対話的にパッケージを選択してインポート（自動検出時） |

**対応行**: `tap "user/repo"`、`brew "formula"`、`cask "name"`、`mas "App Name", id: 1234567890`。vscode / go / cargo / flatpak 等はスキップ（`--verbose` で確認可能）。
