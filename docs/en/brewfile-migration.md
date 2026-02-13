# Brewfile migration

Import an existing Brewfile into al. By default only registration is performed (you bring the current environment under al’s control).

## Preparation

- Create a target profile: `al profile add <profile>`.
- Add providers as needed: `al provider add brew`, `al provider add mas` depending on the Brewfile.

## From a Brewfile

```bash
al import Brewfile --prf core --dry-run   # preview
al import Brewfile --prf core             # register only
al import Brewfile --prf core --install   # also install missing packages
```

## Auto-detect from current brew/mas (no Brewfile)

```bash
al import --prf core --dry-run   # see what would be registered
al import --prf core            # register from current state
al import --prf core -i         # interactive choice of packages
```

**Note**: Auto-detect only includes explicitly installed packages (`brew leaves`). Dependency-only packages are excluded. Cask and mas are treated as explicit.

## Options

| Option | Description |
|--------|-------------|
| `--profile`, `--prf` | Target profile (required). |
| `-s`, `--stage` | Stage name. |
| `--dry-run` | Show parse result and would-be registrations only; no writes. |
| `--install` | Install missing packages via brew/mas (Brewfile mode). |
| `--overwrite` | Overwrite existing entry with same id/provider/profile. |
| `--verbose` | Explain why lines were skipped (Brewfile mode). |
| `-i`, `--interactive` | Choose packages interactively (auto-detect mode). |

## Supported lines

`tap "user/repo"`, `brew "formula"`, `cask "name"`, `mas "App Name", id: 1234567890`. vscode, go, cargo, flatpak, etc. are skipped; use `--verbose` to see why.
