# Concepts

## Profile and Stage

- **Profile**: A unit of environment (e.g. core, work, private). Each profile has its own package list. Use `al add --prf <profile> ...` to target one.
- **Stage**: Whether a package is trial (testing) or stable (production). Full names look like `profile_name.stage_name` (e.g. `core.trial`, `core.stable`). `al init` creates `core` with stage trial. Add more profiles with `al profile add`.

## Promote

If a profile has a `promote_to` set, you can move packages from trial to that target (e.g. stable). Use `al package move <name> --to <profile>` or the alias `al promote <name>` (moves to the current profile’s promote_to). A package (same ID and provider) cannot exist in both trial and stable at once.

## Trial review deadline

If a profile has `review_days`, its packages become subject to review. When a package’s `review_by` is missing or past, it is “expired”. `al activate` prints expired items to stderr and suggests `al review`. In `al review` you can **remove**, **promote**, or **postpone** (extend `review_by` by review_days). Postpone may ask for confirmation.

## Provider

Where packages come from: **brew** (Homebrew formula/cask), **mas** (Mac App Store), **manual** (tracked only, no install). `al init` adds brew. Add mas with `al provider add mas`. Default provider: `al config set --default-provider`.

Dependencies between providers are in `providers.json` (`depends_on`). By default **mas depends on brew**; `al provider add mas` and `al sync` run in dependency order (brew then mas).

**brew tap**: Taps are managed separately under the brew provider (`~/.al/brew-taps.json`). They don’t appear in `al list`. Adding e.g. `al add homebrew/cask-fonts` registers the tap. Update taps with `al provider upgrade` or `al upgrade`. `al provider prune` untaps taps that have no `owner/repo/toolname` packages and removes them from brew-taps (homebrew/core and homebrew/cask are never removed).

## link.d

Store real files/dirs under `~/.al/link.d/<name>/content` and make the path you use (e.g. `~/.config/foo`) a symlink to that content. Sync/backup keep things in sync across machines and GitHub. Add with `al link add <name> <user_path>`. To attach to a package: `al package link add/remove/edit` (link name is usually the package name).

## shell.d

Per-package shell config lives under `~/.al/shell.d/<package-id>/`. `al activate` outputs code to source the enabled ones. Order is controlled with `after`. Enable/disable with `al shell enable/disable`.

## bootstrap

One-off script for setup that doesn’t fit elsewhere. Stored as `~/.al/bootstrap/script.sh`. Create with `al bootstrap add`, edit with `edit`, remove with `remove`, show with `show`. Sync/backup share it across machines.

## extends

A profile can extend another (e.g. `work` extends `core.stable`). Then `al sync --profile work` applies both work and core.stable packages.

## sync and backup

- **sync**: If `~/.al` doesn’t exist, clone the given `owner/repo`, then apply providers, packages, and link.d symlinks. If it exists, only apply. Finally run `~/.al/bootstrap/script.sh` if present. Use `--all` for all AutoSync profiles, `--profile <name>` for one profile and its extends. `--pkg-only` / `--link-only` limit to packages or links. **`--prv <name>`** syncs only packages from a specific provider (e.g. `--prv brew`). **`--plan`** runs a dry-run (no changes; shows what would change and upgradeable packages; use `AL_DEBUG=1` for debug logs). Manual-provider packages trigger a warning to install them yourself.
- **backup**: Commit `~/.al` and push to GitHub. `--init` creates the repo if needed. `--repo owner/repo` sets the target; default is the user’s `dotal` from `gh`. **`--dry-run`** shows what would be backed up without committing or pushing.
