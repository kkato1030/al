package provider

import (
	"fmt"

	"github.com/kkato1030/al/internal/provider"
	"github.com/spf13/cobra"
)

// NewProviderAddCmd creates the provider add command
func NewProviderAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add <provider-name>",
		Short: "Add a provider",
		Long:  "Add and install a package manager provider",
		Args:  cobra.ExactArgs(1),
		RunE:  runProviderAdd,
	}
}

func runProviderAdd(cmd *cobra.Command, args []string) error {
	providerName := args[0]
	var p provider.Provider

	switch providerName {
	case "brew":
		p = provider.NewBrewProvider()
	case "mas":
		p = provider.NewMasProvider()
	case "manual":
		p = provider.NewManualProvider()
	default:
		return fmt.Errorf("unknown provider: %s\nAvailable providers: brew, mas, manual", providerName)
	}

	// Check if already installed
	installed, err := p.CheckInstalled()
	if err != nil {
		return fmt.Errorf("error checking installation: %w", err)
	}

	if installed {
		fmt.Printf("%s is already installed\n", providerName)
		// Still set up config in case it's not configured
		if err := p.SetupConfig(); err != nil {
			fmt.Printf("Warning: failed to set up config: %v\n", err)
		}
		return nil
	}

	// Install the provider
	fmt.Printf("Installing %s...\n", providerName)
	if err := p.Install(); err != nil {
		return fmt.Errorf("error installing %s: %w", providerName, err)
	}

	// Set up config
	fmt.Printf("Setting up configuration for %s...\n", providerName)
	if err := p.SetupConfig(); err != nil {
		return fmt.Errorf("error setting up config: %w", err)
	}

	fmt.Printf("%s has been successfully installed and configured\n", providerName)
	return nil
}
