package provider

import (
	"github.com/spf13/cobra"
)

// NewProviderCmd creates the provider command
func NewProviderCmd() *cobra.Command {
	providerCmd := &cobra.Command{
		Use:   "provider",
		Short: "Manage providers",
		Long:  "Manage package manager providers",
	}

	providerCmd.AddCommand(NewProviderAddCmd())
	providerCmd.AddCommand(NewProviderListCmd())

	return providerCmd
}
