# Profile と Provider

## Profile

- **一覧**: `al profile list`
- **追加**: `al profile add [名前]`（`-t` でテンプレートから作成）
- **詳細**: `al profile show [名前]`
- **削除**: `al profile remove <名前>`
- **テンプレート**: `al profile template` で利用可能なテンプレート一覧

Profile は用途別のパッケージセット（例: work / private）を表す。追加・一覧では `--prf`/`--profile`、適用では `al sync --profile <名前>` を使う。**extends** で別の profile を継承できる（例: work が core.stable を extends）。

## Provider

- **一覧**: `al provider list`
- **追加**: `al provider add <名前>`（brew, mas など）。依存関係（例: mas → brew）も自動で先に処理される。
- **アップグレード**: `al provider upgrade [名前]` — 名前省略で全 provider。brew の場合は `brew update` が実行される。
- **Prune**（brew のみ）: `al provider prune` で、パッケージ一覧に含まれない tap を untap し brew-taps から削除。`--dry-run` で確認、`-y` で確認スキップ。homebrew/core と homebrew/cask は対象外。

デフォルト provider: `al config set --default-provider`。
