# Link and shell

## link.d

Store dotfiles under `~/.al/link.d/<name>/content` and make the path you use (e.g. `~/.config/foo`) a symlink. Sync/backup keep it in sync.

- **Add**: `al link add <name> <user_path>`
- **List**: `al link list`
- **Remove**: `al link remove <name>` (optionally copy content back to user path)
- **Edit**: `al link edit <name>`

To attach a link to a package (e.g. name = package name): `al package link add/remove/edit`.

## shell.d

Per-package shell snippets under `~/.al/shell.d/<package-id>/` are sourced by `al activate`. Order is controlled with `after`.

- **Show**: `al shell show <package>`
- **Add**: `al shell add <package>` — create and edit
- **Edit**: `al shell edit <package>` (must exist; use add to create)
- **Remove**: `al shell remove <package>`
- **Enable/disable**: `al shell enable <package>`, `al shell disable <package>` (disable keeps the file)

Add `eval "$(al activate zsh)"` (or bash) to `.zshrc`/`.bashrc` so this runs in your shell.
