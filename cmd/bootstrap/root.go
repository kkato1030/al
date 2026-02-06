package bootstrap

import (
	"github.com/spf13/cobra"
)

// NewBootstrapCmd creates the bootstrap command
func NewBootstrapCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bootstrap",
		Short: "Manage bootstrap script for new PC setup",
		Long:  "One-off shell script for new PC setup. Script is stored at ~/.al/bootstrap/script.sh. Use add to create, edit to open in EDITOR, remove to delete, show to print content.",
	}
	cmd.AddCommand(NewAddCmd())
	cmd.AddCommand(NewEditCmd())
	cmd.AddCommand(NewRemoveCmd())
	cmd.AddCommand(NewShowCmd())
	return cmd
}
