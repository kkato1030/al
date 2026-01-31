package link

import (
	"github.com/spf13/cobra"
)

// NewLinkCmd creates the link command
func NewLinkCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "link",
		Short: "Manage link.d (symlinked config files and dirs)",
		Long:  "Add, list, remove, or edit link.d entries. User path becomes a symlink to ~/.al/link.d/<id>/content.",
	}
	cmd.AddCommand(NewAddCmd())
	cmd.AddCommand(NewListCmd())
	cmd.AddCommand(NewRemoveCmd())
	cmd.AddCommand(NewEditCmd())
	return cmd
}
