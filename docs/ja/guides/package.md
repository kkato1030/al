# パッケージ管理

- **追加**: `al add <名前>` または `al package add <名前>`。`--prf`、`--prv`、`-s` で profile/provider/stage を指定。mas の場合は `al add --prv mas "App Name" --id <id>`。
- **削除**: `al remove <名前>` または `al package remove <名前>`。
- **一覧**: `al list` または `al package list`。`--profile` で絞り込み。brew tap は一覧に出ない（provider 側で管理）。
- **昇格**: trial から stable（または profile の promote_to）へ: `al promote <名前>` または `al package move <名前> --to <profile>`。
- **詳細**: `al package show <名前>`。
- **検索**: `al package search <検索語>`。
- **アップグレード**: `al package upgrade` でパッケージのみ。`al upgrade` で全 provider と全パッケージ。
- **取り込み**: `al import [Brewfile] --prf <profile>` — [Brewfile からの移行](../brewfile-migration.md) を参照。
- **shell/link**: `al shell *`、`al package link *` — [Link と Shell](link-shell.md) を参照。

エイリアス一覧: `al config alias list`。
