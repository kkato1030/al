# コマンドリファレンス

## ルートコマンド

| コマンド | 説明 |
|----------|------|
| `al init` | 初回セットアップ（profile core, provider brew, デフォルト設定）。`--guided` で対話的なガイド付きセットアップ（推奨） |
| `al activate zsh` / `bash` | シェル用コードを出力（Homebrew 初期化、brew/mas のフック、shell.d のスニペット）。`.zshrc` 等に `eval "$(al activate zsh)"` を追加する。trial のレビュー期限切れがあると stderr に案内を表示 |
| `al review` | レビュー期限切れの trial パッケージを対話で解決（remove / promote / postpone） |
| `al doctor` | 環境の破損や不整合を検出（プロバイダの有無、設定ファイルの妥当性、壊れた symlink、shell.d の依存サイクル、期限切れパッケージ、無効なプロファイル参照など）。システムに変更は加えない。OK / WARN / ERROR でステータスを表示 |
| `al sync [owner/repo]` | `~/.al` が無ければ clone し、provider/パッケージ/link を適用。最後に bootstrap スクリプトがあれば実行。`--plan` で変更内容をプレビュー。`--all` で全プロファイル、`--profile <name>` で特定プロファイルを対象にできる。`--prv <name>` で特定プロバイダのパッケージのみを同期（例: `--prv brew`） |
| `al diff` | 現在のシステムの状態とプロファイルで定義された期待する状態を比較。追加（+）、削除（-）、アップグレード（~）を表示。差分がある場合は終了コード 1 を返す |
| `al backup` | `~/.al` を commit して GitHub に push（`--init` でリポジトリ作成、`--dry-run` でプレビュー）。`--pull` でリモートバックアップの変更を `~/.al` にフェッチ・マージ（コンフリクト時は解決方法を案内） |
| `al update` | al 本体を最新版に更新 |
| `al upgrade` | 全 provider と全パッケージをアップグレード（`-y` で確認スキップ） |
| `al logs` | 実行ログの一覧表示・閲覧。`--list` で最近のログを一覧、引数にログファイル名を指定すると閲覧。ログは `~/.al/logs/` に YYYYMMDD-HHMMSS.log 形式で保存される |
| `al version` | バージョン表示 |

## エイリアス（`al config alias list` で一覧）

`al add` → `al package add`、`al remove` → `al package remove`、`al list` → `al package list`、`al promote` → trial から promote_to への移動、`al import` → `al package import`、`al pkg` → `al package`、`al prf` → `al profile`、`al prv` → `al provider`

## al config

| サブコマンド | 説明 |
|--------------|------|
| `al config set` | `--default-provider`, `--default-profile`, `--default-stage`, `--backup-repo` を設定 |
| `al config show` | 現在の設定を表示 |
| `al config alias list` | エイリアス一覧 |

## al profile

| サブコマンド | 説明 |
|--------------|------|
| `al profile add [名前]` | profile を追加（`-t` でテンプレートから作成）。`--review-days <n>` で既存 profile のレビュー期間（日数）を設定・更新 |
| `al profile list` | 一覧 |
| `al profile show [名前]` | 詳細 |
| `al profile remove <名前>` | 削除 |
| `al profile template` | 利用可能なテンプレート一覧 |

## al provider

| サブコマンド | 説明 |
|--------------|------|
| `al provider add <名前>` | 登録（brew / mas など）。依存関係も自動で解決して先に処理 |
| `al provider list` | 一覧 |
| `al provider upgrade [provider-name]` | provider のアップグレード（未指定時は全 provider）。brew の場合は `brew update` |
| `al provider prune` | brew のみ: 未使用 tap を untap し brew-taps から削除。`--dry-run` で確認、`-y` で確認スキップ |

## al package

| サブコマンド | 説明 |
|--------------|------|
| `al package add [名前]` | 追加（`--provider`, `--profile`, `--stage`, `--id` など） |
| `al package list` | 一覧（`--profile` で絞り込み）。brew tap は表示しない（provider 側で管理） |
| `al package show <名前>` | 詳細 |
| `al package remove <名前>` | 削除 |
| `al package move <名前> --to <profile>` | 別 profile へ移動 |
| `al package import [Brewfile]` | Brewfile から取り込み、または brew/mas から自動検出してインポート |
| `al package search <検索語>` | 検索 |
| `al package upgrade` | 登録パッケージのアップグレード |
| `al shell` | shell.d の show/add/edit/remove/enable/disable |
| `al package link` | link.d の add/remove/edit |

## al link

| サブコマンド | 説明 |
|--------------|------|
| `al link add <名前> <ユーザパス>` | link.d に登録し、ユーザパスを symlink に |
| `al link list` | 一覧 |
| `al link remove <名前>` | 削除（オプションで実体をユーザパスへ copy-back） |
| `al link edit <名前>` | ユーザパスなどの編集 |

## al shell

パッケージごとのシェルスニペット（`~/.al/shell.d/<パッケージ識別子>/`）を管理。

| サブコマンド | 説明 |
|--------------|------|
| `al shell show <パッケージ名>` | シェルスニペットの内容と設定を表示 |
| `al shell add <パッケージ名>` | 新規シェルスニペットを作成し、エディタで編集 |
| `al shell edit <パッケージ名>` | 既存シェルスニペットを編集（未作成時はエラー。`add` を利用） |
| `al shell remove <パッケージ名>` | シェルスニペットを削除 |
| `al shell enable <パッケージ名>` | `al activate` での読み込みを有効化 |
| `al shell disable <パッケージ名>` | `al activate` での読み込みを無効化（ファイルは保持） |

## al bootstrap

新規 PC のセットアップで、ワンオフのシェル実行用。スクリプトは `~/.al/bootstrap/script.sh` に保存される。

| サブコマンド | 説明 |
|--------------|------|
| `al bootstrap add` | スクリプトを作成（未作成時のみ初期内容で作成） |
| `al bootstrap edit` | EDITOR でスクリプトを編集（未作成時は add と同様に作成してから開く） |
| `al bootstrap remove` | スクリプトを削除 |
| `al bootstrap show` | スクリプト内容を表示 |

## al logs

`al sync` や `al upgrade` の実行ログを管理。ログは `~/.al/logs/` に YYYYMMDD-HHMMSS.log 形式で保存される。

| 用法 | 説明 |
|------|------|
| `al logs` | 最新のログを開く |
| `al logs --list` | 最近のログファイル一覧（`-n` で件数指定、デフォルト 10） |
| `al logs <ファイル名>` | 特定のログファイルを開く |

**ログローテーション**: 最新 30 件のみ保持。ログディレクトリは `~/.al/.gitignore` に自動で追加されるため、`al backup` で GitHub に push されない。
