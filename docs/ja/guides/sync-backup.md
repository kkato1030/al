# Sync と Backup

## sync

- **初回（`~/.al` が無い）**: `al sync owner/repo` でリポジトリを `~/.al` に clone し、provider・パッケージ・link.d の symlink を適用。最後に `~/.al/bootstrap/script.sh` があれば実行する。
- **既存の `~/.al`**: `al sync [owner/repo]` は適用のみ（clone しない）。
- **対象**: `--all` で全 AutoSync プロファイル。`--profile <名前>` で指定プロファイルとその extends のみ。
- **限定**: `--pkg-only` または `--link-only` でパッケージのみ / link のみを適用。
- **ドライラン**: `al sync --plan` で変更内容（およびアップグレード可能パッケージ）を表示し、適用は行わない。`AL_DEBUG=1` でデバッグログを有効化。
- **JSON**: `al sync --plan --json` で機械可読なプラン出力。

manual provider のパッケージはインストールされない。al はインストールを促す警告を表示する。

## backup

- **push**: `al backup` で `~/.al` を commit して GitHub に push。
- **リポジトリ作成**: `al backup --init` でリポジトリが無ければ作成してから push。
- **保存先**: `--repo owner/repo` でバックアップ先を指定。デフォルトは `gh` で取得したユーザの `dotal`。
- **ドライラン**: `al backup --dry-run` で commit・push せずに何が backup されるか表示。

`~/.al/logs/` は backup の対象外（.gitignore に含まれる）。
