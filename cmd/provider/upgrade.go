package provider

import (
	"fmt"
	"strings"

	"github.com/kkato1030/al/internal/config"
	"github.com/kkato1030/al/internal/provider"
	"github.com/spf13/cobra"
)

// NewProviderUpgradeCmd creates the provider upgrade command
func NewProviderUpgradeCmd() *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "upgrade [provider-name]",
		Short: "Upgrade provider(s)",
		Long:  "Upgrade a specific provider or all providers. If provider-name is not provided, all providers will be upgraded.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return runProviderUpgradeAll(yes)
			}
			return runProviderUpgrade(args[0])
		},
	}

	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}

// RunProviderUpgradeAll upgrades all providers
func RunProviderUpgradeAll(yes bool) error {
	return runProviderUpgradeAll(yes)
}

func runProviderUpgradeAll(yes bool) error {
	// Load providers config
	providersConfig, err := config.LoadProvidersConfig()
	if err != nil {
		return fmt.Errorf("error loading providers config: %w", err)
	}

	if len(providersConfig.Providers) == 0 {
		fmt.Println("No providers found.")
		return nil
	}

	// Ask for confirmation
	if !yes {
		fmt.Printf("This will upgrade all %d provider(s):\n", len(providersConfig.Providers))
		for _, p := range providersConfig.Providers {
			fmt.Printf("  - %s", p.Name)
			if p.Version != "" {
				fmt.Printf(" (current version: %s)", p.Version)
			}
			fmt.Println()
		}
		fmt.Print("\nDo you want to continue? [y/N]: ")
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			fmt.Println("Upgrade cancelled.")
			return nil
		}
	}

	// Upgrade each provider
	for _, providerConfig := range providersConfig.Providers {
		fmt.Printf("\nUpgrading provider: %s\n", providerConfig.Name)
		if err := runProviderUpgrade(providerConfig.Name); err != nil {
			fmt.Printf("Error upgrading %s: %v\n", providerConfig.Name, err)
			continue
		}
	}

	fmt.Println("\nAll providers upgrade completed.")
	return nil
}

func runProviderUpgrade(providerName string) error {
	// Validate provider exists
	providerConfig, err := config.GetProvider(providerName)
	if err != nil {
		return fmt.Errorf("error loading provider: %w", err)
	}
	if providerConfig == nil {
		return fmt.Errorf("provider '%s' does not exist", providerName)
	}

	// Get provider instance
	var p provider.Provider
	switch providerName {
	case "brew":
		p = provider.NewBrewProvider()
	case "mas":
		p = provider.NewMasProvider()
	default:
		return fmt.Errorf("unknown provider: %s\nAvailable providers: brew, mas", providerName)
	}

	// Check if provider is installed
	installed, err := p.CheckInstalled()
	if err != nil {
		return fmt.Errorf("error checking installation: %w", err)
	}
	if !installed {
		return fmt.Errorf("provider '%s' is not installed", providerName)
	}

	// Upgrade the provider
	if err := p.Upgrade(); err != nil {
		return fmt.Errorf("error upgrading %s: %w", providerName, err)
	}

	return nil
}
