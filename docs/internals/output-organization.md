# Output Organization

This document describes the output organization improvements in `al`.

## Overview

The `al` CLI now distinguishes between its own output and output from internal tools (brew, mas, etc.). This makes the CLI output cleaner and easier to understand.

## Features

### 1. Consistent Output Formatting

`al` uses consistent formatting with icons across all commands:

- **✓** Success messages (green)
- **⚠️** Warning messages (yellow)
- **✗** Error messages (red)
- **ℹ** Information messages

### 2. Tool Output Suppression

By default, `al` suppresses stdout from internal tools (brew, mas, shell commands) and only shows its own messages. However, stderr is always displayed to ensure password prompts and error messages are visible:

```bash
$ al package add git
Installing package git (formula:git)...
✓ Installed git
```

If a tool requires sudo password, the prompt will be visible even in normal mode:
```bash
$ al package add some-tool
Installing package some-tool...
Password:  # This prompt is visible and interactive
✓ Installed some-tool
```

### 3. Debug Mode

Set `AL_DEBUG=1` to see all output including internal tool output with prefixes:

```bash
$ AL_DEBUG=1 al package add git
[DEBUG 10:52:11.123] Checking if provider brew is installed
Installing package git (formula:git)...
[brew] ==> Downloading git...
[brew] ==> Pouring git-2.x.x.bottle.tar.gz
[brew] 🍺  /usr/local/Cellar/git/2.x.x: 1,234 files, 56.7MB
✓ Installed git
```

Debug mode also shows:
- Timestamps for debug messages
- Internal tool output with `[tool]` prefix
- All debug logging from al itself

### 4. Loading Indicators

Long-running operations show a spinner to indicate progress:

```bash
$ al package add firefox
⠹ Installing firefox...
✓ Installed firefox
```

Spinners are automatically disabled in debug mode so you can see the actual tool output.

## Environment Variables

### AL_DEBUG

Set `AL_DEBUG=1` to enable debug mode:

```bash
# Normal mode - clean output
al sync

# Debug mode - verbose output
AL_DEBUG=1 al sync
```

## Implementation Details

### Output Package

The `internal/output` package provides utilities for consistent output:

- `Info()` - Print information messages
- `Success()` - Print success messages with ✓
- `Warning()` - Print warning messages with ⚠️
- `Error()` - Print error messages with ✗
- `DebugLog()` - Print debug messages (only when AL_DEBUG=1)
- `NewSpinner()` - Create a loading spinner
- `GetToolOutputWriter()` - Get appropriate writer for tool stdout
- `GetToolErrorWriter()` - Get appropriate writer for tool stderr (always shows stderr for password prompts)

### Provider Updates

Provider implementations (brew, mas) have been updated to:

1. Use spinners for long-running operations
2. Suppress tool output by default
3. Show tool output with prefix when AL_DEBUG=1
4. Use consistent success/error messages

### Example Code

```go
import "github.com/kkato1030/al/internal/output"

// Show a message
output.Info("Installing package %s...", pkgName)

// Use a spinner for long operation
spinner := output.NewSpinner(fmt.Sprintf("Installing %s...", pkgName))
spinner.Start()

// Execute command with appropriate output handling
cmd := exec.Command("brew", "install", pkgName)
cmd.Stdin = os.Stdin // Always connect stdin for interactive prompts
cmd.Stdout = output.GetToolOutputWriter("brew") // Suppress stdout unless in debug mode
cmd.Stderr = output.GetToolErrorWriter("brew")  // Always show stderr for password prompts
err := cmd.Run()

spinner.Stop()

// Show result
if err != nil {
    return fmt.Errorf("failed: %w", err)
}
output.Success("Installed %s", pkgName)
```

## Benefits

1. **Cleaner Output**: Users see only what they need to know
2. **Better UX**: Loading indicators show progress
3. **Easy Debugging**: Debug mode shows everything for troubleshooting
4. **Consistency**: All commands use the same output style
5. **Professional**: Icons and formatting make output more readable
