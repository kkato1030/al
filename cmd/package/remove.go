package packagecmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kkato1030/al/internal/config"
	"github.com/kkato1030/al/internal/provider"
	"github.com/kkato1030/al/internal/ui"
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

	var foundPkg *config.PackageConfig
	for _, pkg := range packagesConfig.Packages {
		if pkg.Name == packageName && pkg.Provider == providerName && pkg.Profile == profile {
			foundPkg = &pkg
			break
		}
	}

	if foundPkg == nil {
		return fmt.Errorf("package '%s' with provider '%s' in profile '%s' not found", packageName, providerName, profile)
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

	// Uninstall the package using ID
	if err := p.UninstallPackage(foundPkg.ID); err != nil {
		return fmt.Errorf("error uninstalling package: %w", err)
	}

	// Remove the package from config
	if err := config.RemovePackage(foundPkg.ID, providerName, profile); err != nil {
		return fmt.Errorf("error removing package: %w", err)
	}

	fmt.Printf("Package '%s' (ID: %s) has been successfully removed from profile '%s' with provider '%s'\n", packageName, foundPkg.ID, profile, providerName)
	return nil
}

func runPackageRemoveInteractive(packageName, provider, profile string) error {
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

	// Multiple matches, let user select with UI
	model := ui.NewPackageSelectModel(matchingPackages, fmt.Sprintf("Select package to remove (found %d matching packages)", len(matchingPackages)))
	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running UI: %w", err)
	}

	selectedPkg := model.GetSelected()
	if selectedPkg == nil {
		return fmt.Errorf("package selection is required")
	}

	return runPackageRemove(packageName, selectedPkg.Provider, selectedPkg.Profile)
}
