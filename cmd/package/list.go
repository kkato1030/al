package packagecmd

import (
	"fmt"
	"sort"
	"strings"

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

	// Group packages by profile and provider
	// Structure: map[profile]map[provider][]PackageConfig
	grouped := make(map[string]map[string][]config.PackageConfig)
	for _, pkg := range filteredPackages {
		profileName := pkg.Profile
		if profileName == "" {
			profileName = "(no profile)"
		}
		providerName := pkg.Provider
		if providerName == "" {
			providerName = "(no provider)"
		}

		if grouped[profileName] == nil {
			grouped[profileName] = make(map[string][]config.PackageConfig)
		}
		grouped[profileName][providerName] = append(grouped[profileName][providerName], pkg)
	}

	// Sort profiles and providers for consistent output
	profiles := make([]string, 0, len(grouped))
	for profile := range grouped {
		profiles = append(profiles, profile)
	}
	sort.Strings(profiles)

	fmt.Println("Configured packages:")
	for i, profileName := range profiles {
		if i > 0 {
			fmt.Println()
		}
		fmt.Printf("%s\n", profileName)

		providers := make([]string, 0, len(grouped[profileName]))
		for provider := range grouped[profileName] {
			providers = append(providers, provider)
		}
		sort.Strings(providers)

		for _, providerName := range providers {
			packages := grouped[profileName][providerName]
			// Sort packages by name for consistent output
			sort.Slice(packages, func(i, j int) bool {
				return packages[i].Name < packages[j].Name
			})

			// Build comma-separated list of package names
			packageNames := make([]string, len(packages))
			for idx, pkg := range packages {
				packageNames[idx] = pkg.Name
			}

			fmt.Printf("  %s: %s\n", providerName, strings.Join(packageNames, ", "))
		}
	}

	return nil
}
