# Command Reference

## Root commands

| Command | Description |
|---------|-------------|
| `al init` | First-time setup (profile core, provider brew, defaults). Use `--guided` for interactive setup (recommended). |
| `al activate zsh` / `bash` | Output shell code (Homebrew init, brew/mas hooks, shell.d snippets). Add `eval "$(al activate zsh)"` to `.zshrc` etc. Shows trial review expiry on stderr when applicable. |
| `al review` | Interactively resolve expired trial packages (remove / promote / postpone). |
| `al doctor` | Check for issues (providers, config, broken symlinks, shell.d cycles, expired packages, invalid profile refs). No system changes. Shows OK / WARN / ERROR. |
| `al sync [owner/repo]` | Clone if `~/.al` missing, then apply providers/packages/links; run bootstrap if present. `--plan` to preview without applying (checks upgrades too). `AL_DEBUG=1` for debug logs. `--all` or `--profile <name>`. `--prv <name>` to sync only packages from a specific provider (e.g. `--prv brew`). |
| `al diff` | Compare current system state to profiles. Shows additions (+), removals (-), upgrades (~). Exit code 1 if there is drift. |
| `al backup` | Commit `~/.al` and push to GitHub. `--init` to create repo, `--dry-run` to preview. `--pull` to fetch changes from remote (deprecated in favor of `al pull`). |
| `al pull` | Fetch and merge changes from the remote backup into `~/.al`. Use `--repo` to override the repository, `--dry-run` to preview. Conflicts are reported with resolution guidance. |
| `al update` | Update al to the latest release. |
| `al upgrade` | Upgrade all providers and packages. Use `-y` to skip confirmations. |
| `al logs` | View logs. `--list` for recent log files; pass a filename to open one. Logs live in `~/.al/logs/` as YYYYMMDD-HHMMSS.log. |
| `al version` | Show version. |

## Aliases (`al config alias list`)

`al add` → `al package add`, `al remove` → `al package remove`, `al list` → `al package list`, `al promote` → move to profile’s promote_to, `al import` → `al package import`, `al pkg` → `al package`, `al prf` → `al profile`, `al prv` → `al provider`.

## al config

| Subcommand | Description |
|------------|-------------|
| `al config set` | Set `--default-provider`, `--default-profile`, `--default-stage`, `--backup-repo`. |
| `al config show` | Show current config. |
| `al config alias list` | List aliases. |

## al profile

| Subcommand | Description |
|------------|-------------|
| `al profile add [name]` | Add profile. Use `-t` to create from a template. Use `--review-days <n>` to set (or update) the review period in days for an existing profile. |
| `al profile list` | List profiles. |
| `al profile show [name]` | Show details. |
| `al profile remove <name>` | Remove profile. |
| `al profile template` | List available templates. |

## al provider

| Subcommand | Description |
|------------|-------------|
| `al provider add <name>` | Register (e.g. brew, mas). Dependencies are resolved and applied first. |
| `al provider list` | List providers. |
| `al provider upgrade [name]` | Upgrade provider(s). Omit name to upgrade all. For brew, runs `brew update`. |
| `al provider prune` | brew only: untap taps that have no `owner/repo/toolname` packages and remove from brew-taps. homebrew/core and homebrew/cask are excluded. `--dry-run` to preview, `-y` to skip confirm. |

## al package

| Subcommand | Description |
|------------|-------------|
| `al package add [name]` | Add package (`--provider`, `--profile`, `--stage`, `--id`, etc.). |
| `al package list` | List packages (`--profile` to filter). Brew taps are not listed (managed under provider). |
| `al package show <name>` | Show package details. |
| `al package remove <name>` | Remove package. |
| `al package move <name> --to <profile>` | Move to another profile. |
| `al package import [Brewfile]` | Import from Brewfile or auto-detect from brew/mas (`--prf`, `--install`, `--dry-run`, `-i`, etc.). |
| `al package search <query>` | Search. |
| `al package upgrade` | Upgrade registered packages. |
| `al shell` | shell.d: show/add/edit/remove/enable/disable. |
| `al package link` | link.d: add/remove/edit. |

## al link

| Subcommand | Description |
|------------|-------------|
| `al link add <name> <user_path>` | Register in link.d and make user path a symlink. |
| `al link list` | List links. |
| `al link remove <name>` | Remove (optionally copy content back to user path). |
| `al link edit <name>` | Edit user path etc. |

## al shell

Manage per-package shell snippets under `~/.al/shell.d/<package-id>/`.

| Subcommand | Description |
|------------|-------------|
| `al shell show <package>` | Show snippet content and config. |
| `al shell add <package>` | Create and edit a new snippet. |
| `al shell edit <package>` | Edit existing (fails if none; use add). |
| `al shell remove <package>` | Remove snippet. |
| `al shell enable <package>` | Enable for `al activate`. |
| `al shell disable <package>` | Disable (file kept). |

## al bootstrap

One-off script at `~/.al/bootstrap/script.sh` for setup that doesn’t fit elsewhere.

| Subcommand | Description |
|------------|-------------|
| `al bootstrap add` | Create script (only if missing). |
| `al bootstrap edit` | Edit with EDITOR (creates if missing). |
| `al bootstrap remove` | Remove script. |
| `al bootstrap show` | Show script content. |

## al logs

Logs for `al sync` and `al upgrade` are in `~/.al/logs/` as YYYYMMDD-HHMMSS.log (command, timestamp, stdout, stderr).

| Usage | Description |
|-------|-------------|
| `al logs` | Open latest log. |
| `al logs --list` | List recent logs (`-n` for count, default 10). |
| `al logs <filename>` | Open specific log (e.g. `al logs 20260212-123456.log`). |

**Rotation**: Only the latest 30 logs are kept. The logs directory is in `~/.al/.gitignore`, so `al backup` does not push logs.
