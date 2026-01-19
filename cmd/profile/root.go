package profile

import (
	"github.com/spf13/cobra"
)

// NewProfileCmd creates the profile command
func NewProfileCmd() *cobra.Command {
	profileCmd := &cobra.Command{
		Use:   "profile",
		Short: "Manage profiles",
		Long:  "Manage profiles for configuration",
	}

	profileCmd.AddCommand(NewProfileAddCmd())
	profileCmd.AddCommand(NewProfileListCmd())
	profileCmd.AddCommand(NewProfileShowCmd())
	profileCmd.AddCommand(NewProfileRemoveCmd())

	return profileCmd
}
