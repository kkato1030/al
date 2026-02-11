package provider

import (
	"fmt"
	"strings"

	"github.com/kkato1030/al/internal/config"
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

	orderedProviders, err := config.ResolveProvidersWithDependencies([]string{providerName})
	if err != nil {
		return fmt.Errorf("error resolving provider dependencies: %w", err)
	}

	if len(orderedProviders) > 1 {
		fmt.Printf("Resolved provider dependencies: %s\n", strings.Join(orderedProviders, " -> "))
	}

	for _, name := range orderedProviders {
		if err := addProvider(name); err != nil {
			return err
		}
	}

	return nil
}

func addProvider(providerName string) error {
	p, err := providerFromName(providerName)
	if err != nil {
		return err
	}

	installed, err := p.CheckInstalled()
	if err != nil {
		return fmt.Errorf("error checking installation for %s: %w", providerName, err)
	}

	if installed {
		fmt.Printf("%s is already installed\n", providerName)
		if err := p.SetupConfig(); err != nil {
			fmt.Printf("Warning: failed to set up config for %s: %v\n", providerName, err)
		}
		return nil
	}

	fmt.Printf("Installing %s...\n", providerName)
	if err := p.Install(); err != nil {
		return fmt.Errorf("error installing %s: %w", providerName, err)
	}

	fmt.Printf("Setting up configuration for %s...\n", providerName)
	if err := p.SetupConfig(); err != nil {
		return fmt.Errorf("error setting up config for %s: %w", providerName, err)
	}

	fmt.Printf("%s has been successfully installed and configured\n", providerName)
	return nil
}

func providerFromName(providerName string) (provider.Provider, error) {
	switch providerName {
	case "brew":
		return provider.NewBrewProvider(), nil
	case "mas":
		return provider.NewMasProvider(), nil
	case "manual":
		return provider.NewManualProvider(), nil
	default:
		return nil, fmt.Errorf("unknown provider: %s\nAvailable providers: brew, mas, manual", providerName)
	}
}
