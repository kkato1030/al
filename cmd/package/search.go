package packagecmd

import (
	"bufio"
	"fmt"
	"os"

	"github.com/kkato1030/al/internal/config"
	"github.com/kkato1030/al/internal/provider"
	"github.com/spf13/cobra"
)

// NewPackageSearchCmd creates the package search command
func NewPackageSearchCmd() *cobra.Command {
	var providerName string

	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search for packages",
		Long:  "Search for packages using the specified provider. If provider is not specified, interactive mode will be used.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := args[0]

			// If provider is not specified, use interactive mode
			if providerName == "" {
				return runPackageSearchInteractive(query)
			}

			return runPackageSearch(query, providerName)
		},
	}

	cmd.Flags().StringVarP(&providerName, "provider", "p", "", "Provider name (required)")

	return cmd
}

func runPackageSearch(query, providerName string) error {
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
		return fmt.Errorf("unsupported provider: %s", providerName)
	}

	// Search for packages
	results, err := p.SearchPackage(query)
	if err != nil {
		return fmt.Errorf("error searching packages: %w", err)
	}

	if len(results) == 0 {
		fmt.Printf("No packages found for query '%s' with provider '%s'\n", query, providerName)
		return nil
	}

	// Display results
	fmt.Printf("\nFound %d package(s) for query '%s' with provider '%s':\n\n", len(results), query, providerName)
	for i, result := range results {
		fmt.Printf("  %d. %s", i+1, result.Name)
		if result.ID != "" {
			fmt.Printf(" (ID: %s)", result.ID)
		}
		if result.Description != "" {
			fmt.Printf(" - %s", result.Description)
		}
		fmt.Println()
	}

	return nil
}

func runPackageSearchInteractive(query string) error {
	scanner := bufio.NewScanner(os.Stdin)

	// Get provider
	selectedProvider, err := selectProvider(scanner)
	if err != nil {
		return err
	}
	if selectedProvider == "" {
		return fmt.Errorf("provider is required")
	}

	return runPackageSearch(query, selectedProvider)
}
