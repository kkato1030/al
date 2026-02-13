# Bootstrap と Logs

## Bootstrap

他コマンドでは自動化できないワンオフのセットアップ用スクリプト。`~/.al/bootstrap/script.sh` に 1 本保存。sync/backup で他マシンと共有可能。

- **追加**: `al bootstrap add` — 未作成時のみ作成
- **編集**: `al bootstrap edit` — EDITOR で開く（無ければ作成してから開く）
- **削除**: `al bootstrap remove`
- **表示**: `al bootstrap show`

`al sync` の最後にこのスクリプトがあれば実行される。

## Logs

`al sync` と `al upgrade` の実行ログは `~/.al/logs/` に YYYYMMDD-HHMMSS.log 形式で保存される（実行コマンド・タイムスタンプ・標準出力・標準エラー出力）。

- **最新**: `al logs` で最新のログを開く
- **一覧**: `al logs --list`（`-n` で件数、デフォルト 10）
- **指定**: `al logs <ファイル名>`（例: `al logs 20260212-123456.log`）

最新 30 件のみ保持。ログディレクトリは `~/.al/.gitignore` に含まれるため、`al backup` では push されない。
