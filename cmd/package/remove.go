package packagecmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/kkato1030/al/internal/config"
	"github.com/kkato1030/al/internal/provider"
	"github.com/spf13/cobra"
)

// NewPackageRemoveCmd creates the package remove command
func NewPackageRemoveCmd() *cobra.Command {
	var provider string
	var profile string

	cmd := &cobra.Command{
		Use:   "remove <package-name>",
		Short: "Remove a package",
		Long:  "Remove a package from a profile. If required flags are not provided, interactive mode will be used.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			packageName := args[0]

			// If required flags are not set, use interactive mode
			if provider == "" || profile == "" {
				return runPackageRemoveInteractive(packageName, provider, profile)
			}

			return runPackageRemove(packageName, provider, profile)
		},
	}

	cmd.Flags().StringVarP(&provider, "provider", "p", "", "Provider name (required)")
	cmd.Flags().StringVarP(&profile, "profile", "f", "", "Profile name (required)")

	return cmd
}

func runPackageRemove(packageName, providerName, profile string) error {
	// Check if package exists
	packagesConfig, err := config.LoadPackagesConfig()
	if err != nil {
		return fmt.Errorf("error loading packages config: %w", err)
	}

	found := false
	for _, pkg := range packagesConfig.Packages {
		if pkg.Name == packageName && pkg.Provider == providerName && pkg.Profile == profile {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("package '%s' with provider '%s' in profile '%s' not found", packageName, providerName, profile)
	}

	// Get provider instance
	var p provider.Provider
	switch providerName {
	case "brew":
		p = provider.NewBrewProvider()
	default:
		return fmt.Errorf("unsupported provider: %s", providerName)
	}

	// Uninstall the package
	if err := p.UninstallPackage(packageName); err != nil {
		return fmt.Errorf("error uninstalling package: %w", err)
	}

	// Remove the package from config
	if err := config.RemovePackage(packageName, providerName, profile); err != nil {
		return fmt.Errorf("error removing package: %w", err)
	}

	fmt.Printf("Package '%s' has been successfully removed from profile '%s' with provider '%s'\n", packageName, profile, providerName)
	return nil
}

func runPackageRemoveInteractive(packageName, provider, profile string) error {
	scanner := bufio.NewScanner(os.Stdin)

	// Get package name
	fmt.Printf("Package name: %s\n", packageName)

	// Load packages config to find matching packages
	packagesConfig, err := config.LoadPackagesConfig()
	if err != nil {
		return fmt.Errorf("error loading packages config: %w", err)
	}

	// Find packages matching the name
	var matchingPackages []config.PackageConfig
	for _, pkg := range packagesConfig.Packages {
		if pkg.Name == packageName {
			// If provider is specified, filter by it
			if provider != "" && pkg.Provider != provider {
				continue
			}
			// If profile is specified, filter by it
			if profile != "" && pkg.Profile != profile {
				continue
			}
			matchingPackages = append(matchingPackages, pkg)
		}
	}

	if len(matchingPackages) == 0 {
		return fmt.Errorf("package '%s' not found", packageName)
	}

	// If only one match, use it directly
	if len(matchingPackages) == 1 {
		pkg := matchingPackages[0]
		fmt.Printf("Found package: %s (provider: %s, profile: %s)\n", pkg.Name, pkg.Provider, pkg.Profile)
		return runPackageRemove(packageName, pkg.Provider, pkg.Profile)
	}

	// Multiple matches, let user select
	fmt.Printf("\nFound %d matching packages:\n", len(matchingPackages))
	for i, pkg := range matchingPackages {
		fmt.Printf("  %d. %s (provider: %s, profile: %s", i+1, pkg.Name, pkg.Provider, pkg.Profile)
		if pkg.Version != "" {
			fmt.Printf(", version: %s", pkg.Version)
		}
		if pkg.Description != "" {
			fmt.Printf(" - %s", pkg.Description)
		}
		fmt.Println(")")
	}
	fmt.Print("Select package to remove (number): ")

	if !scanner.Scan() {
		return fmt.Errorf("failed to read input")
	}

	input := strings.TrimSpace(scanner.Text())
	if input == "" {
		return fmt.Errorf("package selection is required")
	}

	idx, err := strconv.Atoi(input)
	if err != nil {
		return fmt.Errorf("invalid number: %s", input)
	}

	if idx < 1 || idx > len(matchingPackages) {
		return fmt.Errorf("number %d is out of range (1-%d)", idx, len(matchingPackages))
	}

	selectedPkg := matchingPackages[idx-1]
	return runPackageRemove(packageName, selectedPkg.Provider, selectedPkg.Profile)
}
