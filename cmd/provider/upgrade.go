package provider

import (
	"fmt"
	"os"

	"github.com/kkato1030/al/internal/config"
	"github.com/kkato1030/al/internal/output"
	"github.com/kkato1030/al/internal/prompt"
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
		fmt.Println("No providers configured")
		return nil
	}

	providerVersions := make(map[string]string, len(providersConfig.Providers))
	var providerNames []string
	for _, p := range providersConfig.Providers {
		providerVersions[p.Name] = p.Version
		providerNames = append(providerNames, p.Name)
	}

	orderedProviders, err := config.OrderProvidersByDependency(providerNames)
	if err != nil {
		return fmt.Errorf("error resolving provider upgrade order: %w", err)
	}

	// Ask for confirmation
	if !yes {
		fmt.Printf("This will upgrade all %d provider(s):\n", len(orderedProviders))
		for _, name := range orderedProviders {
			fmt.Printf("  - %s", name)
			if providerVersions[name] != "" {
				fmt.Printf(" (current version: %s)", providerVersions[name])
			}
			fmt.Println()
		}
		ok, err := prompt.Confirm(os.Stderr, "\nContinue? [y/N]: ")
		if err != nil {
			return err
		}
		if !ok {
			fmt.Fprintln(os.Stderr, "Cancelled.")
			return nil
		}
	}

	// Upgrade each provider
	for _, providerName := range orderedProviders {
		output.Info("\nUpgrading provider: %s", providerName)
		if err := runProviderUpgrade(providerName); err != nil {
			output.Error("Upgrading %s: %v", providerName, err)
			continue
		}
	}

	output.Success("All providers upgraded")
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
