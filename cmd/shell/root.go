package shell

import (
	"github.com/spf13/cobra"
)

// NewCmd creates the shell command
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "shell",
		Short: "Manage shell.d snippets",
		Long:  "Show, add, edit, remove, enable, or disable shell.d snippets for packages (used by al activate).",
	}
	cmd.AddCommand(NewShowCmd())
	cmd.AddCommand(NewAddCmd())
	cmd.AddCommand(NewEditCmd())
	cmd.AddCommand(NewRemoveCmd())
	cmd.AddCommand(NewEnableCmd())
	cmd.AddCommand(NewDisableCmd())
	return cmd
}
