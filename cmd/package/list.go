package packagecmd

import (
	"fmt"

	"github.com/kkato1030/al/internal/config"
	"github.com/spf13/cobra"
)

// NewPackageListCmd creates the package list command
func NewPackageListCmd() *cobra.Command {
	var profile string
	var provider string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all packages",
		Long:  "List all configured packages. Optionally filter by profile and/or provider.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPackageList(profile, provider)
		},
	}

	cmd.Flags().StringVarP(&profile, "profile", "f", "", "Filter packages by profile name")
	cmd.Flags().StringVarP(&provider, "provider", "p", "", "Filter packages by provider name")

	return cmd
}

func runPackageList(profileFilter, providerFilter string) error {
	packagesConfig, err := config.LoadPackagesConfig()
	if err != nil {
		return fmt.Errorf("error loading packages config: %w", err)
	}

	// Filter packages based on provided flags
	var filteredPackages []config.PackageConfig
	for _, pkg := range packagesConfig.Packages {
		matchesProfile := profileFilter == "" || pkg.Profile == profileFilter
		matchesProvider := providerFilter == "" || pkg.Provider == providerFilter

		if matchesProfile && matchesProvider {
			filteredPackages = append(filteredPackages, pkg)
		}
	}

	if len(filteredPackages) == 0 {
		if profileFilter != "" || providerFilter != "" {
			fmt.Println("No packages found matching the specified filters")
		} else {
			fmt.Println("No packages configured")
		}
		return nil
	}

	fmt.Println("Configured packages:")
	for _, pkg := range filteredPackages {
		fmt.Printf("  - %s", pkg.Name)
		if pkg.Provider != "" {
			fmt.Printf(" (provider: %s", pkg.Provider)
			if pkg.Profile != "" {
				fmt.Printf(", profile: %s", pkg.Profile)
			}
			fmt.Printf(")")
		} else if pkg.Profile != "" {
			fmt.Printf(" (profile: %s)", pkg.Profile)
		}
		if pkg.Version != "" {
			fmt.Printf(" [version: %s]", pkg.Version)
		}
		if pkg.Description != "" {
			fmt.Printf(" - %s", pkg.Description)
		}
		fmt.Println()
	}

	return nil
}
