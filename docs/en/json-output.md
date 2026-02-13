# JSON output

Use the global `--json` flag for machine-readable output and CI/CD.

## Commands that support `--json`

- **`al diff --json`**: Additions, removals, and upgrades as JSON.
- **`al doctor --json`**: Diagnostic result and per-check status as JSON.
- **`al sync --plan --json`**: Planned actions and summary as JSON.

## Example: `al diff --json`

```json
{
  "additions": [
    {"type": "addition", "provider": "brew", "name": "jq", "id": "formula:jq"}
  ],
  "removals": [],
  "upgrades": [],
  "has_drift": true
}
```

## Behavior

- Exit codes are unchanged: e.g. `al diff` still exits 1 when there is drift, `al doctor` exits 1 on errors.
- Human-readable output is unchanged when `--json` is not used.
- JSON is written only to stdout (no headers or extra text when `--json` is set).
