# Bootstrap and logs

## Bootstrap

One-off script for setup that doesn’t fit other commands. Stored as `~/.al/bootstrap/script.sh`. Sync/backup share it across machines.

- **Add**: `al bootstrap add` — create script (only if it doesn’t exist).
- **Edit**: `al bootstrap edit` — open in EDITOR (creates if missing).
- **Remove**: `al bootstrap remove`
- **Show**: `al bootstrap show`

`al sync` runs this script at the end if it exists.

## Logs

`al sync` and `al upgrade` write logs to `~/.al/logs/` as YYYYMMDD-HHMMSS.log (command, timestamp, stdout, stderr).

- **Latest**: `al logs` opens the most recent log.
- **List**: `al logs --list` (use `-n` for count, default 10).
- **Specific**: `al logs <filename>` (e.g. `al logs 20260212-123456.log`).

Only the latest 30 logs are kept. The logs directory is in `~/.al/.gitignore`, so it is not pushed by `al backup`.
