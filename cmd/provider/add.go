package provider

import (
	"fmt"
	"strings"

	"github.com/kkato1030/al/internal/config"
	"github.com/kkato1030/al/internal/output"
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
		output.Info("Resolved provider dependencies: %s", strings.Join(orderedProviders, " -> "))
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
		output.Info("%s is already installed", providerName)
		if err := p.SetupConfig(); err != nil {
			output.Warning("Failed to set up config for %s: %v", providerName, err)
		}
		return nil
	}

	output.Info("Installing %s...", providerName)
	if err := p.Install(); err != nil {
		return fmt.Errorf("error installing %s: %w", providerName, err)
	}

	output.Info("Setting up configuration for %s...", providerName)
	if err := p.SetupConfig(); err != nil {
		return fmt.Errorf("error setting up config for %s: %w", providerName, err)
	}

	output.Success("Installed %s", providerName)
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
