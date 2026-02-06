package cmd

import (
	"fmt"
	"os"

	"github.com/kkato1030/al/internal/config"
	"github.com/spf13/cobra"
)

// NewActivateCmd creates the activate command
func NewActivateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "activate [shell]",
		Short: "Output shell code to source shell.d snippets and brew/mas hooks",
		Long:  "Output shell code to source enabled shell.d snippets in topological order and to wrap brew/mas for install/uninstall (prompts to use al). Add eval \"$(al activate zsh)\" to your .zshrc (al does not edit .zshrc).",
		Args:  cobra.ExactArgs(1),
		RunE:  runActivate,
	}
	return cmd
}

func runActivate(cmd *cobra.Command, args []string) error {
	// Show overdue packages notice on stderr (so eval still gets clean script on stdout)
	if overdue, err := config.GetOverduePackages(); err == nil && len(overdue) > 0 {
		fmt.Fprintln(os.Stderr, "Review overdue: the following packages need a decision (remove / promote / postpone).")
		for _, pkg := range overdue {
			fmt.Fprintf(os.Stderr, "  - %s (profile: %s, provider: %s)\n", pkg.Name, pkg.Profile, pkg.Provider)
		}
		fmt.Fprintln(os.Stderr, "Run 'al review' to resolve.")
	}

	shell := args[0]
	ext, err := shellExt(shell)
	if err != nil {
		return err
	}
	// Output brew/mas hook first (wraps install/uninstall to suggest al)
	if hook := brewMasHook(shell); hook != "" {
		fmt.Print(hook)
	}
	entries, err := config.GetEnabledShellEntriesInOrder(ext)
	if err != nil {
		return err
	}
	for _, e := range entries {
		for _, p := range e.Paths {
			fmt.Printf("source %q\n", p)
		}
	}
	return nil
}

func brewMasHook(shell string) string {
	switch shell {
	case "zsh":
		return brewMasHookZsh
	case "bash":
		return brewMasHookBash
	default:
		return ""
	}
}

// brewMasHookZsh wraps brew and mas for zsh (install/uninstall → suggest al, [y/N] to run directly).
const brewMasHookZsh = `
_al_hook_first_subcmd() {
  local arg
  for arg in "$@"; do
    case "$arg" in
      -*) ;;
      *) echo "$arg"; return ;;
    esac
  done
}

_al_hook_brew_intercept() {
  case "$(_al_hook_first_subcmd "$@")" in
    install|uninstall) return 0 ;;
    *) return 1 ;;
  esac
}

_al_hook_mas_intercept() {
  case "${1:-}" in
    install|uninstall) return 0 ;;
    *) return 1 ;;
  esac
}

brew() {
  if _al_hook_brew_intercept "$@"; then
    echo "For install/uninstall, consider using al: al add <package> or al remove <package>"
    read -q "?Run brew directly? [y/N] " || true
    echo
    if [[ "$REPLY" != [yY] ]]; then
      echo "Skipped. Use 'al add <package>' or 'al remove <package>' to manage packages."
      return 0
    fi
  fi
  command brew "$@"
}

mas() {
  if _al_hook_mas_intercept "$@"; then
    echo "For install/uninstall, consider using al: al add <package> or al remove <package>"
    read -q "?Run mas directly? [y/N] " || true
    echo
    if [[ "$REPLY" != [yY] ]]; then
      echo "Skipped. Use 'al add <package>' or 'al remove <package>' to manage packages."
      return 0
    fi
  fi
  command mas "$@"
}
`

// brewMasHookBash wraps brew and mas for bash (install/uninstall → suggest al, [y/N] to run directly).
const brewMasHookBash = `
_al_hook_first_subcmd() {
  local arg
  for arg in "$@"; do
    case "$arg" in
      -*) ;;
      *) echo "$arg"; return ;;
    esac
  done
}

_al_hook_brew_intercept() {
  case "$(_al_hook_first_subcmd "$@")" in
    install|uninstall) return 0 ;;
    *) return 1 ;;
  esac
}

_al_hook_mas_intercept() {
  case "${1:-}" in
    install|uninstall) return 0 ;;
    *) return 1 ;;
  esac
}

brew() {
  if _al_hook_brew_intercept "$@"; then
    echo "For install/uninstall, consider using al: al add <package> or al remove <package>"
    read -n 1 -p "Run brew directly? [y/N] " _al_hook_reply
    echo
    if [[ "${_al_hook_reply}" != [yY] ]]; then
      echo "Skipped. Use 'al add <package>' or 'al remove <package>' to manage packages."
      return 0
    fi
  fi
  command brew "$@"
}

mas() {
  if _al_hook_mas_intercept "$@"; then
    echo "For install/uninstall, consider using al: al add <package> or al remove <package>"
    read -n 1 -p "Run mas directly? [y/N] " _al_hook_reply
    echo
    if [[ "${_al_hook_reply}" != [yY] ]]; then
      echo "Skipped. Use 'al add <package>' or 'al remove <package>' to manage packages."
      return 0
    fi
  fi
  command mas "$@"
}
`

func shellExt(shell string) (string, error) {
	switch shell {
	case "zsh":
		return ".zsh", nil
	case "bash":
		return ".bash", nil
	default:
		return "", fmt.Errorf("unsupported shell: %s (use zsh or bash)", shell)
	}
}
