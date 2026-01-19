package main

import (
	"fmt"
	"os"

	"github.com/kkato1030/al/internal/provider"
	"github.com/spf13/cobra"
)

var version = "0.1.0"

func main() {
	rootCmd := &cobra.Command{
		Use:   "al",
		Short: "Mac Management Tools",
		Long:  "al - Mac Management Tools",
	}

	rootCmd.AddCommand(newVersionCmd())
	rootCmd.AddCommand(newProviderCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version number",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("al version %s\n", version)
		},
	}
}

func newProviderCmd() *cobra.Command {
	providerCmd := &cobra.Command{
		Use:   "provider",
		Short: "Manage providers",
		Long:  "Manage package manager providers",
	}

	providerCmd.AddCommand(newProviderAddCmd())

	return providerCmd
}

func newProviderAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add <provider-name>",
		Short: "Add a provider",
		Long:  "Add and install a package manager provider",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			providerName := args[0]
			var p provider.Provider

			switch providerName {
			case "brew":
				p = provider.NewBrewProvider()
			default:
				return fmt.Errorf("unknown provider: %s\nAvailable providers: brew", providerName)
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
		},
	}
}
