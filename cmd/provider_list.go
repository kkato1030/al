package cmd

import (
	"fmt"

	"github.com/kkato1030/al/internal/config"
	"github.com/spf13/cobra"
)

// NewProviderListCmd creates the provider list command
func NewProviderListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all providers",
		Long:  "List all installed package manager providers",
		RunE:  runProviderList,
	}
}

func runProviderList(cmd *cobra.Command, args []string) error {
	config, err := config.LoadProvidersConfig()
	if err != nil {
		return fmt.Errorf("error loading providers config: %w", err)
	}

	if len(config.Providers) == 0 {
		fmt.Println("No providers installed")
		return nil
	}

	fmt.Println("Installed providers:")
	for _, p := range config.Providers {
		fmt.Printf("  - %s", p.Name)
		if p.Version != "" {
			fmt.Printf(" (version: %s)", p.Version)
		}
		if !p.InstalledAt.IsZero() {
			fmt.Printf(" (installed at: %s)", p.InstalledAt.Format("2006-01-02 15:04:05"))
		}
		fmt.Println()
	}

	return nil
}
