package shell

import (
	"github.com/spf13/cobra"
)

// NewCmd creates the shell subcommand
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "shell",
		Short: "Manage package shell.d snippets",
		Long:  "Show, set, unset, edit, enable, or disable shell.d snippets for a package (used by al activate).",
	}
	cmd.AddCommand(NewShowCmd())
	cmd.AddCommand(NewSetCmd())
	cmd.AddCommand(NewUnsetCmd())
	cmd.AddCommand(NewEditCmd())
	cmd.AddCommand(NewEnableCmd())
	cmd.AddCommand(NewDisableCmd())
	return cmd
}
