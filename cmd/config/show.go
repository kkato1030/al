package config

import (
	"fmt"

	"github.com/kkato1030/al/internal/config"
	"github.com/spf13/cobra"
)

// NewConfigShowCmd creates the config show command
func NewConfigShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		Long:  "Display the current configuration values",
		RunE: func(cmd *cobra.Command, args []string) error {
			appConfig, err := config.LoadAppConfig()
			if err != nil {
				return fmt.Errorf("error loading config: %w", err)
			}

			fmt.Println("Current configuration:")
			if appConfig.DefaultProvider != "" {
				fmt.Printf("  default_provider: %s\n", appConfig.DefaultProvider)
			} else {
				fmt.Println("  default_provider: (not set)")
			}

			if appConfig.DefaultProfile != "" {
				fmt.Printf("  default_profile: %s\n", appConfig.DefaultProfile)
			} else {
				fmt.Println("  default_profile: (not set)")
			}

			if appConfig.DefaultStage != "" {
				fmt.Printf("  default_stage: %s\n", appConfig.DefaultStage)
			} else {
				fmt.Println("  default_stage: (not set)")
			}

			return nil
		},
	}
}
