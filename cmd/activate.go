package cmd

import (
	"fmt"

	"github.com/kkato1030/al/internal/config"
	"github.com/spf13/cobra"
)

// NewActivateCmd creates the activate command
func NewActivateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "activate [shell]",
		Short: "Output shell code to source shell.d snippets",
		Long:  "Output shell code to source enabled shell.d snippets in topological order. Add eval \"$(al activate zsh)\" to your .zshrc (al does not edit .zshrc).",
		Args:  cobra.ExactArgs(1),
		RunE:  runActivate,
	}
	return cmd
}

func runActivate(cmd *cobra.Command, args []string) error {
	shell := args[0]
	ext, err := shellExt(shell)
	if err != nil {
		return err
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
