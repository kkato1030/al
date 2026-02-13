# Package management

- **Add**: `al add <name>` or `al package add <name>`. Use `--prf`, `--prv`, `-s` for profile/provider/stage. For mas: `al add --prv mas "App Name" --id <id>`.
- **Remove**: `al remove <name>` or `al package remove <name>`.
- **List**: `al list` or `al package list`; use `--profile` to filter. Brew taps are not listed (see provider).
- **Promote**: Move from trial to stable (or profile’s promote_to): `al promote <name>` or `al package move <name> --to <profile>`.
- **Show**: `al package show <name>`.
- **Search**: `al package search <query>`.
- **Upgrade**: `al package upgrade` for packages; `al upgrade` for all providers and packages.
- **Import**: `al import [Brewfile] --prf <profile>` — see [Brewfile migration](../brewfile-migration.md).
- **Shell/link**: `al shell *`, `al package link *` — see [Link & Shell](link-shell.md).

Aliases: `al config alias list`.
