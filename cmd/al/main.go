package main

import (
	"fmt"
	"os"

	"github.com/kkato1030/al/internal/provider"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("al - Mac Management Tools")
		fmt.Println("Usage: al <command> [options]")
		os.Exit(1)
	}

	command := os.Args[1]
	switch command {
	case "version":
		fmt.Println("al version 0.1.0")
	case "provider":
		handleProviderCommand()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}

func handleProviderCommand() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: al provider <subcommand> [options]")
		fmt.Println("Subcommands:")
		fmt.Println("  add <provider-name>  Add a provider")
		os.Exit(1)
	}

	subcommand := os.Args[2]
	switch subcommand {
	case "add":
		handleProviderAdd()
	default:
		fmt.Printf("Unknown provider subcommand: %s\n", subcommand)
		fmt.Println("Usage: al provider <subcommand> [options]")
		fmt.Println("Subcommands:")
		fmt.Println("  add <provider-name>  Add a provider")
		os.Exit(1)
	}
}

func handleProviderAdd() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: al provider add <provider-name>")
		fmt.Println("Available providers: brew")
		os.Exit(1)
	}

	providerName := os.Args[3]
	var p provider.Provider

	switch providerName {
	case "brew":
		p = provider.NewBrewProvider()
	default:
		fmt.Printf("Unknown provider: %s\n", providerName)
		fmt.Println("Available providers: brew")
		os.Exit(1)
	}

	// Check if already installed
	installed, err := p.CheckInstalled()
	if err != nil {
		fmt.Printf("Error checking installation: %v\n", err)
		os.Exit(1)
	}

	if installed {
		fmt.Printf("%s is already installed\n", providerName)
		// Still set up config in case it's not configured
		if err := p.SetupConfig(); err != nil {
			fmt.Printf("Warning: failed to set up config: %v\n", err)
		}
		return
	}

	// Install the provider
	fmt.Printf("Installing %s...\n", providerName)
	if err := p.Install(); err != nil {
		fmt.Printf("Error installing %s: %v\n", providerName, err)
		os.Exit(1)
	}

	// Set up config
	fmt.Printf("Setting up configuration for %s...\n", providerName)
	if err := p.SetupConfig(); err != nil {
		fmt.Printf("Error setting up config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("%s has been successfully installed and configured\n", providerName)
}
