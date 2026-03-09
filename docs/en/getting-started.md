# Getting Started

## Install

**Latest**

```bash
curl -fsSL https://raw.githubusercontent.com/kkato1030/al/main/install.sh | bash
al version
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

## Initialize

```bash
al init
```

This creates the config directory `~/.al` (or `$AL_HOME`), profile `core` (stage: trial), registers provider `brew`, and saves defaults (profile=core, provider=brew, stage=trial).

### Guided init (recommended)

For first-time setup, use the interactive guided init:

```bash
al init --guided
```

You’ll be asked:

1. **Profile setup**: Single profile (core only) or multiple profiles (core + extra)
2. **Extra profile names**: If multiple, comma-separated (e.g. work, personal)
3. **Trial workflow**: Whether to use a trial period for new packages
4. **Review period**: If trial is enabled, review deadline in days (1/7/14/30/60)

## Add packages

```bash
al add jq
al add --prv mas "Xcode" --id 497799835
```

Default is brew. Use `--profile`/`--prf`, `--provider`/`--prv`, `--stage`/`-s` to target a different profile/provider/stage. Uninstalled packages are installed when added.

## Enable shell integration

Add one line to `.zshrc` or `.bashrc` (al does not edit your shell config):

```bash
eval "$(al activate zsh)"   # zsh
eval "$(al activate bash)"  # bash
```

This sets up Homebrew (`brew shellenv`), sources enabled shell.d snippets, and hooks for `brew install/uninstall` and `mas install/uninstall` (with a prompt to use al). Works on Apple Silicon (`/opt/homebrew`) and Intel (`/usr/local`).

## Sync and backup

- Apply `~/.al` from a GitHub repo: `al sync owner/repo` (clone on first run, then apply only).
- Push current config to GitHub: `al backup --init` (creates the repo if needed and pushes).

Next: [Concepts](concepts.md) | [Command Reference](command-reference.md)
