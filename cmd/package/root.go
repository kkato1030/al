package packagecmd

import (
	"github.com/spf13/cobra"
)

// NewPackageCmd creates the package command
func NewPackageCmd() *cobra.Command {
	packageCmd := &cobra.Command{
		Use:   "package",
		Short: "Manage packages",
		Long:  "Manage packages for profiles and providers",
	}

	packageCmd.AddCommand(NewPackageAddCmd())
	packageCmd.AddCommand(NewPackageListCmd())
	packageCmd.AddCommand(NewPackageShowCmd())
	packageCmd.AddCommand(NewPackageRemoveCmd())
	packageCmd.AddCommand(NewPackageMoveCmd())
	packageCmd.AddCommand(NewPackageSearchCmd())

	return packageCmd
}
