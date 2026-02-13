# Workflow examples

## Check environment (after setup or when something’s wrong)

```bash
al doctor
```

Example output:

```
[OK]    brew is available
[WARN]  link ~/.config/ghostty is broken
[WARN]  trial package review expired: gh (run 'al review')
```

## Trial to stable (promote)

1. `al add <package>` to add to trial.
2. After trying it: `al promote <package>` to promote, or `al remove <package>` if you don’t need it.

## Separate work and personal

1. `al profile add work -e core.stable -p core.stable`
2. `al profile add private`
3. `al add <package> --prf work` or `al add <package> --prf private`
4. `al sync --profile work` or `al sync --all` to apply.

## Preview changes (plan mode)

1. `al sync --plan` to see what would change (no apply).
2. If it looks good, run `al sync` to apply.

## Share and restore via GitHub

1. `al backup --init` to create the repo and push.
2. On another machine: `al sync owner/dotal` to clone and apply.
3. After changes: `al backup` to push.
