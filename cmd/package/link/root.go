package link

import (
	"github.com/spf13/cobra"
)

// NewCmd creates the package link subcommand (alias for al link --package).
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "link",
		Short: "Manage link.d for a package (link name = package name)",
		Long:  "Add, remove, or edit the link.d entry for a package. Link name is the package name; one link per package.",
	}
	cmd.AddCommand(NewAddCmd())
	cmd.AddCommand(NewRemoveCmd())
	cmd.AddCommand(NewEditCmd())
	return cmd
}
