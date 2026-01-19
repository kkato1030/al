package config

import (
	"fmt"

	"github.com/kkato1030/al/internal/config"
	"github.com/spf13/cobra"
)

// NewConfigSetCmd creates the config set command
func NewConfigSetCmd() *cobra.Command {
	var defaultProvider string
	var defaultProfile string

	cmd := &cobra.Command{
		Use:   "set",
		Short: "Set configuration values",
		Long:  "Set default_provider and/or default_profile",
		RunE: func(cmd *cobra.Command, args []string) error {
			if defaultProvider == "" && defaultProfile == "" {
				return fmt.Errorf("at least one of --default-provider or --default-profile must be specified")
			}

			if defaultProvider != "" {
				if err := config.SetDefaultProvider(defaultProvider); err != nil {
					return fmt.Errorf("error setting default provider: %w", err)
				}
				fmt.Printf("Default provider set to: %s\n", defaultProvider)
			}

			if defaultProfile != "" {
				if err := config.SetDefaultProfile(defaultProfile); err != nil {
					return fmt.Errorf("error setting default profile: %w", err)
				}
				fmt.Printf("Default profile set to: %s\n", defaultProfile)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&defaultProvider, "default-provider", "", "Set the default provider")
	cmd.Flags().StringVar(&defaultProfile, "default-profile", "", "Set the default profile")

	return cmd
}
