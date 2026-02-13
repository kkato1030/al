package packagecmd

import (
	"fmt"
	"os"

	"github.com/kkato1030/al/internal/config"
	"github.com/kkato1030/al/internal/output"
	"github.com/kkato1030/al/internal/prompt"
	"github.com/kkato1030/al/internal/provider"
	"github.com/spf13/cobra"
)

// NewPackageUpgradeCmd creates the package upgrade command
func NewPackageUpgradeCmd() *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "upgrade [package-name]",
		Short: "Upgrade package(s)",
		Long:  "Upgrade a specific package or all packages. If package-name is not provided, all packages will be upgraded.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return runPackageUpgradeAll(yes)
			}
			return runPackageUpgrade(args[0])
		},
	}

	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}

// RunPackageUpgradeAll upgrades all packages
func RunPackageUpgradeAll(yes bool) error {
	return runPackageUpgradeAll(yes)
}

func runPackageUpgradeAll(yes bool) error {
	// Load packages config
	packagesConfig, err := config.LoadPackagesConfig()
	if err != nil {
		return fmt.Errorf("error loading packages config: %w", err)
	}

	if len(packagesConfig.Packages) == 0 {
		fmt.Println("No packages configured")
		return nil
	}

	// Ask for confirmation
	if !yes {
		fmt.Printf("This will upgrade all %d package(s):\n", len(packagesConfig.Packages))
		for _, pkg := range packagesConfig.Packages {
			fmt.Printf("  - %s (%s:%s)", pkg.Name, pkg.Provider, pkg.ID)
			if pkg.Profile != "" {
				fmt.Printf(" [profile: %s]", pkg.Profile)
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

	// Group packages by provider for efficient upgrading
	packagesByProvider := make(map[string][]config.PackageConfig)
	for _, pkg := range packagesConfig.Packages {
		packagesByProvider[pkg.Provider] = append(packagesByProvider[pkg.Provider], pkg)
	}

	// Upgrade packages by provider
	successCount := 0
	errorCount := 0

	for providerName, packages := range packagesByProvider {
		output.Info("\nUpgrading packages for provider: %s", providerName)

		// Get provider instance
		var p provider.Provider
		switch providerName {
		case "brew":
			p = provider.NewBrewProvider()
		case "mas":
			p = provider.NewMasProvider()
		case "manual":
			p = provider.NewManualProvider()
		default:
			output.Warning("Unknown provider '%s', skipping packages", providerName)
			errorCount += len(packages)
			continue
		}

		// Check if provider is installed
		installed, err := p.CheckInstalled()
		if err != nil {
			output.Error("Checking provider installation: %v", err)
			errorCount += len(packages)
			continue
		}
		if !installed {
			output.Warning("Provider '%s' is not installed, skipping packages", providerName)
			errorCount += len(packages)
			continue
		}

		// Upgrade each package
		for _, pkg := range packages {
			output.Info("Upgrading %s...", pkg.Name)
			if err := p.UpgradePackage(pkg.ID); err != nil {
				output.Error("Upgrading %s: %v", pkg.Name, err)
				errorCount++
			} else {
				successCount++
			}
		}
	}

	output.Success("Upgrade completed: %d succeeded, %d failed", successCount, errorCount)
	return nil
}

func runPackageUpgrade(packageName string) error {
	// Load packages config
	packagesConfig, err := config.LoadPackagesConfig()
	if err != nil {
		return fmt.Errorf("error loading packages config: %w", err)
	}

	// Find packages matching the name
	var matchingPackages []config.PackageConfig
	for _, pkg := range packagesConfig.Packages {
		if pkg.Name == packageName {
			matchingPackages = append(matchingPackages, pkg)
		}
	}

	if len(matchingPackages) == 0 {
		return fmt.Errorf("package '%s' not found", packageName)
	}

	// If multiple packages with same name, upgrade all of them
	if len(matchingPackages) > 1 {
		fmt.Printf("Found %d package(s) with name '%s':\n", len(matchingPackages), packageName)
		for _, pkg := range matchingPackages {
			fmt.Printf("  - %s (%s:%s) [profile: %s]\n", pkg.Name, pkg.Provider, pkg.ID, pkg.Profile)
		}
		fmt.Println("Upgrading all matching packages...")
	}

	// Upgrade each matching package
	for _, pkg := range matchingPackages {
		// Get provider instance
		var p provider.Provider
		switch pkg.Provider {
		case "brew":
			p = provider.NewBrewProvider()
		case "mas":
			p = provider.NewMasProvider()
		case "manual":
			p = provider.NewManualProvider()
		default:
			output.Warning("Unknown provider '%s' for package %s, skipping", pkg.Provider, pkg.Name)
			continue
		}

		// Check if provider is installed
		installed, err := p.CheckInstalled()
		if err != nil {
			output.Error("Checking provider installation: %v", err)
			continue
		}
		if !installed {
			output.Warning("Provider '%s' is not installed for package %s, skipping", pkg.Provider, pkg.Name)
			continue
		}

		// Upgrade the package
		output.Info("Upgrading %s (%s:%s)...", pkg.Name, pkg.Provider, pkg.ID)
		if err := p.UpgradePackage(pkg.ID); err != nil {
			return fmt.Errorf("error upgrading %s: %w", pkg.Name, err)
		}
	}

	return nil
}
