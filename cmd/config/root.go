package config

import (
	"github.com/spf13/cobra"
)

// NewConfigCmd creates the config command
func NewConfigCmd() *cobra.Command {
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration",
		Long:  "Manage application configuration settings",
	}

	configCmd.AddCommand(NewConfigSetCmd())
	configCmd.AddCommand(NewConfigShowCmd())

	return configCmd
}
