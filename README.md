# al — Mac environment manager

[日本語](README.ja.md)

**al** is a Mac environment manager that lets you manage packages (Homebrew / mas) and dotfiles with a **trial → stable** workflow. Use profiles and stages to try new tools safely, then promote them when you’re ready.

---

## Quick Start

**Install** (both paths start here)

```bash
curl -fsSL https://raw.githubusercontent.com/kkato1030/al/main/install.sh | bash
al version
```

Pick the path that matches you.

### Already using Homebrew / mas? Import in one step

If you already manage packages with `brew` (and maybe `mas`), you don't have to start over — bring your **current environment** under al with `al import`. This is the fastest way to migrate.

```bash
al init                        # set up ~/.al
al profile add core            # create a profile to import into
al import --prf core --dry-run # preview what would be registered
al import --prf core           # register your currently installed packages
```

Already have a `Brewfile`? Import it directly:

```bash
al import Brewfile --prf core             # register only
al import Brewfile --prf core --install   # also install missing packages
```

`al import` (auto-detect) only picks up explicitly installed packages (`brew leaves`), plus casks and mas apps. See **[Brewfile migration](docs/en/brewfile-migration.md)** for details.

### Starting from scratch

```bash
al init            # or: al init --guided  for interactive setup
al add jq          # add a package
```

**Enable shell** (both paths) — add to `.zshrc` or `.bashrc`:

```bash
eval "$(al activate zsh)"   # or bash
```

**Sync from GitHub** / **Backup to GitHub**

```bash
al sync owner/repo          # apply config from a repo
al backup --init            # create repo and push (first time)
al pull                     # pull changes from backup into ~/.al
```

---

## Features

| Area | Commands |
|------|----------|
| Packages | `al add`, `al remove`, `al promote`, `al list`, `al upgrade` |
| Diff & review | `al diff`, `al review` (trial deadline) |
| Health | `al doctor` |
| Profiles | `al profile add/list/show` |
| Dotfiles | `al link add/list/remove/edit`, link.d |
| Shell snippets | `al shell show/add/edit/enable/disable`, shell.d |
| Sync / backup | `al sync [owner/repo]`, `al backup`, `al pull` |
| Import | `al import [Brewfile] --prf <profile>` |

Full list and details: **[Documentation (English)](docs/en/README.md)** \| **[ドキュメント（日本語）](docs/ja/README.md)**

---

## Concepts (short)

- **Profile**: An environment unit (e.g. core, work). Each has its own package list. Target with `--prf` / `--profile`.
- **Stage**: **trial** (testing) vs **stable** (production). Promote with `al promote <name>` or `al package move <name> --to <profile>`.
- **Provider**: Where packages come from — **brew**, **mas**, or **manual**. Default: brew.

More: [Concepts (en)](docs/en/concepts.md) \| [概念（ja）](docs/ja/concepts.md)

---

## Command overview

| Command | Description |
|---------|-------------|
| `al init` | First-time setup. `--guided` for interactive. |
| `al activate zsh`/`bash` | Shell integration (add `eval "$(al activate zsh)"` to rc). |
| `al review` | Resolve expired trial packages. |
| `al doctor` | Check environment. |
| `al sync [owner/repo]` | Apply config (clone if needed). `--plan` to preview. `--prv brew` to sync only brew packages. |
| `al diff` | Compare system vs profiles. |
| `al backup` | Commit and push `~/.al` to GitHub. |
| `al pull` | Fetch and merge changes from the remote backup into `~/.al`. |
| `al upgrade` | Upgrade providers and packages. |
| `al config` / `al profile` / `al provider` / `al package` / `al link` / `al shell` / `al bootstrap` / `al logs` | Subcommands. |

Full reference: [docs/en/command-reference.md](docs/en/command-reference.md) \| [docs/ja/command-reference.md](docs/ja/command-reference.md)

---

## Install options

**Latest**

```bash
curl -fsSL https://raw.githubusercontent.com/kkato1030/al/main/install.sh | bash
```

**Specific version**

```bash
AL_VERSION=v1.0.0 curl -fsSL https://raw.githubusercontent.com/kkato1030/al/main/install.sh | bash
```

**Install directory** (default: `/usr/local/bin`)

```bash
AL_INSTALL_DIR=$HOME/bin curl -fsSL https://raw.githubusercontent.com/kkato1030/al/main/install.sh | bash
```

**Self-update**

```bash
al update
```

---

## Development

See [AGENTS.md](AGENTS.md) for build, test, and contribution. Build: `make build`, test: `make test`. e2e tests run on CI (macOS runner).
