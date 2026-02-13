# Sync and backup

## sync

- **First time (no `~/.al`)**: `al sync owner/repo` clones the repo into `~/.al`, then applies providers, packages, and link.d symlinks. If `~/.al/bootstrap/script.sh` exists, it is run at the end.
- **Existing `~/.al`**: `al sync [owner/repo]` only applies (no clone).
- **Scope**: `--all` for all AutoSync profiles; `--profile <name>` for one profile and its extends.
- **Limit**: `--pkg-only` or `--link-only` to apply only packages or only links.
- **Dry-run**: `al sync --plan` shows what would change (and upgradeable packages) without applying. Use `AL_DEBUG=1` for debug logs.
- **JSON**: `al sync --plan --json` for machine-readable plan output.

Manual-provider packages are not installed; al will warn you to install them yourself.

## backup

- **Push**: `al backup` commits `~/.al` and pushes to GitHub.
- **Create repo**: `al backup --init` creates the repo if it doesn’t exist, then pushes.
- **Target**: `--repo owner/repo` to set the backup repo; default is your `dotal` (from `gh`).
- **Dry-run**: `al backup --dry-run` shows what would be backed up without committing or pushing.

Logs under `~/.al/logs/` are ignored by backup (in `.gitignore`).
