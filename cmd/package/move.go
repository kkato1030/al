package packagecmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/kkato1030/al/internal/config"
	"github.com/spf13/cobra"
)

// NewPackageMoveCmd creates the package move command
func NewPackageMoveCmd() *cobra.Command {
	var toProfile string

	cmd := &cobra.Command{
		Use:   "move <package-name>",
		Short: "Move a package to another profile",
		Long:  "Move a package from its current profile to another profile specified with --to flag",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			packageName := args[0]

			if toProfile == "" {
				return fmt.Errorf("--to flag is required to specify the target profile")
			}

			// Verify target profile exists
			targetProfile, err := config.GetProfile(toProfile)
			if err != nil {
				return fmt.Errorf("error loading target profile: %w", err)
			}
			if targetProfile == nil {
				return fmt.Errorf("target profile '%s' does not exist", toProfile)
			}

			return runPackageMove(packageName, toProfile)
		},
	}

	cmd.Flags().StringVar(&toProfile, "to", "", "Target profile name (required)")

	return cmd
}

func runPackageMove(packageName, toProfile string) error {
	// Load packages config to find matching packages
	packagesConfig, err := config.LoadPackagesConfig()
	if err != nil {
		return fmt.Errorf("error loading packages config: %w", err)
	}

	// Find packages matching the name
	var matchingPackages []config.PackageConfig
	for _, pkg := range packagesConfig.Packages {
		if pkg.Name == packageName {
			// Skip if already in target profile
			if pkg.Profile == toProfile {
				return fmt.Errorf("package '%s' is already in profile '%s'", packageName, toProfile)
			}
			matchingPackages = append(matchingPackages, pkg)
		}
	}

	if len(matchingPackages) == 0 {
		return fmt.Errorf("package '%s' not found", packageName)
	}

	// Check if package already exists in target profile with same provider
	for _, pkg := range matchingPackages {
		// Check if same package (name + provider) already exists in target profile
		for _, existingPkg := range packagesConfig.Packages {
			if existingPkg.Name == packageName &&
				existingPkg.Provider == pkg.Provider &&
				existingPkg.Profile == toProfile {
				return fmt.Errorf("package '%s' with provider '%s' already exists in profile '%s'", packageName, pkg.Provider, toProfile)
			}
		}
	}

	// If only one match, use it directly
	if len(matchingPackages) == 1 {
		pkg := matchingPackages[0]
		fmt.Printf("Moving package: %s (provider: %s, from profile: %s, to profile: %s)\n", pkg.Name, pkg.Provider, pkg.Profile, toProfile)
		return movePackage(pkg, toProfile)
	}

	// Multiple matches, let user select
	scanner := bufio.NewScanner(os.Stdin)
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
	fmt.Print("Select package to move (number): ")

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
	fmt.Printf("Moving package: %s (provider: %s, from profile: %s, to profile: %s)\n", selectedPkg.Name, selectedPkg.Provider, selectedPkg.Profile, toProfile)
	return movePackage(selectedPkg, toProfile)
}

func movePackage(pkg config.PackageConfig, toProfile string) error {
	// Remove package from current profile
	if err := config.RemovePackage(pkg.Name, pkg.Provider, pkg.Profile); err != nil {
		return fmt.Errorf("error removing package from current profile: %w", err)
	}

	// Add package to target profile (preserve version, description, and InstalledAt)
	newPkg := config.PackageConfig{
		Name:        pkg.Name,
		Provider:    pkg.Provider,
		Profile:     toProfile,
		Version:     pkg.Version,
		Description: pkg.Description,
		InstalledAt: pkg.InstalledAt,
	}

	if err := config.AddOrUpdatePackage(newPkg); err != nil {
		return fmt.Errorf("error adding package to target profile: %w", err)
	}

	fmt.Printf("Package '%s' has been successfully moved from profile '%s' to profile '%s'\n", pkg.Name, pkg.Profile, toProfile)
	return nil
}
