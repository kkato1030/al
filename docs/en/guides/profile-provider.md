# Profile and provider

## Profiles

- **List**: `al profile list`
- **Add**: `al profile add [name]` (use `-t` for a template).
- **Show**: `al profile show [name]`
- **Remove**: `al profile remove <name>`
- **Templates**: `al profile template` lists available templates.

Profiles define separate package sets (e.g. work vs private). Use `--prf` / `--profile` on add/list and `al sync --profile <name>` to apply a profile. Use **extends** so a profile inherits packages from another (e.g. work extends core.stable).

## Providers

- **List**: `al provider list`
- **Add**: `al provider add <name>` (e.g. brew, mas). Dependencies (e.g. mas → brew) are applied in order.
- **Upgrade**: `al provider upgrade [name]` — omit name to upgrade all. For brew this runs `brew update`.
- **Prune** (brew only): `al provider prune` removes taps that have no packages in your package list. `--dry-run` to preview, `-y` to skip confirm. homebrew/core and homebrew/cask are never removed.

Default provider: `al config set --default-provider`.
