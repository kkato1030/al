package packagecmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kkato1030/al/internal/config"
	"github.com/kkato1030/al/internal/provider"
	"github.com/kkato1030/al/internal/ui"
	"github.com/spf13/cobra"
)

// NewPackageAddCmd creates the package add command
func NewPackageAddCmd() *cobra.Command {
	var provider string
	var profile string
	var version string
	var description string
	var packageID string

	cmd := &cobra.Command{
		Use:   "add [package-name]",
		Short: "Add a package",
		Long:  "Add a package to a profile with a provider. If package-name is not provided, interactive mode will be used.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// If no package name provided, use fully interactive mode
			if len(args) == 0 {
				return runPackageAddInteractive("", provider, profile, version, description, packageID)
			}

			packageName := args[0]

			// Determine provider and profile from flags or defaults
			finalProvider := provider
			finalProfile := profile

			// If flags are not set, try to use defaults
			if finalProvider == "" || finalProfile == "" {
				appConfig, err := config.LoadAppConfig()
				if err != nil {
					return fmt.Errorf("error loading app config: %w", err)
				}

				if finalProvider == "" {
					finalProvider = appConfig.DefaultProvider
				}
				if finalProfile == "" {
					finalProfile = appConfig.DefaultProfile
				}
			}

			// If still not set, return error
			if finalProvider == "" || finalProfile == "" {
				return fmt.Errorf("provider and profile must be specified via flags or default config. Use 'al config set --default-provider <provider> --default-profile <profile>' to set defaults")
			}

			// Verify that provider and profile exist
			providerConfig, err := config.GetProvider(finalProvider)
			if err != nil {
				return fmt.Errorf("error loading provider: %w", err)
			}
			if providerConfig == nil {
				return fmt.Errorf("provider '%s' does not exist", finalProvider)
			}

			profileConfig, err := config.GetProfile(finalProfile)
			if err != nil {
				return fmt.Errorf("error loading profile: %w", err)
			}
			if profileConfig == nil {
				return fmt.Errorf("profile '%s' does not exist", finalProfile)
			}

			return runPackageAdd(packageName, finalProvider, finalProfile, version, description, packageID)
		},
	}

	cmd.Flags().StringVarP(&provider, "provider", "p", "", "Provider name (required)")
	cmd.Flags().StringVarP(&profile, "profile", "f", "", "Profile name (required)")
	cmd.Flags().StringVarP(&version, "version", "v", "", "Package version (optional)")
	cmd.Flags().StringVarP(&description, "description", "d", "", "Package description (optional)")
	cmd.Flags().StringVarP(&packageID, "id", "i", "", "Package ID (required for mas, optional for brew)")

	return cmd
}

// RunPackageAdd runs the package add logic (exported for use by other commands)
func RunPackageAdd(packageName, providerName, profile, version, description, packageID string) error {
	return runPackageAdd(packageName, providerName, profile, version, description, packageID)
}

func runPackageAdd(packageName, providerName, profile, version, description, packageID string) error {
	// Validate provider exists
	providerConfig, err := config.GetProvider(providerName)
	if err != nil {
		return fmt.Errorf("error loading provider: %w", err)
	}
	if providerConfig == nil {
		return fmt.Errorf("provider '%s' does not exist", providerName)
	}

	// Validate profile exists
	profileConfig, err := config.GetProfile(profile)
	if err != nil {
		return fmt.Errorf("error loading profile: %w", err)
	}
	if profileConfig == nil {
		return fmt.Errorf("profile '%s' does not exist", profile)
	}

	// Determine package ID and name based on provider
	var finalID string
	var finalName string
	var p provider.Provider

	switch providerName {
	case "brew":
		brewProvider := provider.NewBrewProvider()
		p = brewProvider
		// For brew, detect package type and generate ID in format "{formula,cask,tap}:<package_name>"
		generatedID, err := brewProvider.GeneratePackageID(packageName)
		if err != nil {
			return fmt.Errorf("error detecting package type: %w", err)
		}
		finalID = generatedID
		finalName = packageName
	case "mas":
		masProvider := provider.NewMasProvider()
		p = masProvider
		// For mas, if --id is not provided, search and let user select
		if packageID == "" {
			// Search for packages
			results, err := masProvider.SearchPackage(packageName)
			if err != nil {
				return fmt.Errorf("error searching packages: %w", err)
			}

			if len(results) == 0 {
				return fmt.Errorf("no packages found for query '%s'", packageName)
			}

			// If only one result, use it automatically
			if len(results) == 1 {
				finalID = results[0].ID
				finalName = results[0].Name
				if finalName == "" {
					finalName = packageName
				}
			} else {
				// Multiple results, let user select with UI
				model := ui.NewSearchResultSelectModel(results, fmt.Sprintf("Select package (found %d package(s) for query '%s')", len(results), packageName))
				p := tea.NewProgram(model)
				if _, err := p.Run(); err != nil {
					return fmt.Errorf("error running UI: %w", err)
				}

				selected := model.GetSelected()
				if selected == nil {
					return fmt.Errorf("package selection is required")
				}

				finalID = selected.ID
				finalName = selected.Name
				if finalName == "" {
					finalName = packageName
				}
			}
		} else {
			// --id is provided
			finalID = packageID
			finalName = packageName
		}
	case "manual":
		manualProvider := provider.NewManualProvider()
		p = manualProvider
		// For manual, use package name as ID
		finalID = packageName
		finalName = packageName
	default:
		return fmt.Errorf("unsupported provider: %s", providerName)
	}

	// Check if package already exists in config
	packagesConfig, err := config.LoadPackagesConfig()
	if err != nil {
		return fmt.Errorf("error loading packages config: %w", err)
	}

	packageExists := false
	for _, existingPkg := range packagesConfig.Packages {
		if existingPkg.ID == finalID && existingPkg.Provider == providerName && existingPkg.Profile == profile {
			packageExists = true
			break
		}
	}

	// Install the package only if it doesn't exist in config
	if !packageExists {
		if err := p.InstallPackage(finalID); err != nil {
			return fmt.Errorf("error installing package: %w", err)
		}
	} else {
		fmt.Printf("Package with id '%s' already exists in config, skipping installation\n", finalID)
	}

	// Create package config
	pkg := config.PackageConfig{
		ID:          finalID,
		Name:        finalName,
		Provider:    providerName,
		Profile:     profile,
		Version:     version,
		Description: description,
	}

	// Add or update package in config
	if err := config.AddOrUpdatePackage(pkg); err != nil {
		return fmt.Errorf("error adding package: %w", err)
	}

	if packageExists {
		fmt.Printf("Package '%s' (ID: %s) has been successfully updated in profile '%s' with provider '%s'\n", finalName, finalID, profile, providerName)
	} else {
		fmt.Printf("Package '%s' (ID: %s) has been successfully added to profile '%s' with provider '%s'\n", finalName, finalID, profile, providerName)
	}
	return nil
}

func runPackageAddInteractive(packageName, provider, profile, version, description, packageID string) error {
	scanner := bufio.NewScanner(os.Stdin)

	// Get package name (if not provided)
	if packageName == "" {
		fmt.Print("Package name: ")
		if !scanner.Scan() {
			return fmt.Errorf("failed to read input")
		}
		packageName = strings.TrimSpace(scanner.Text())
		if packageName == "" {
			return fmt.Errorf("package name is required")
		}
	} else {
		fmt.Printf("Package name: %s\n", packageName)
	}

	// Get provider
	if provider == "" {
		selectedProvider, err := selectProviderUI()
		if err != nil {
			return err
		}
		if selectedProvider == "" {
			return fmt.Errorf("provider is required")
		}
		provider = selectedProvider
	} else {
		fmt.Printf("Provider: %s\n", provider)
	}

	// Get profile
	if profile == "" {
		selectedProfile, err := selectProfileUI()
		if err != nil {
			return err
		}
		if selectedProfile == "" {
			return fmt.Errorf("profile is required")
		}
		profile = selectedProfile
	} else {
		fmt.Printf("Profile: %s\n", profile)
	}

	// Get version
	if version == "" {
		fmt.Print("Version (optional, press Enter to skip): ")
		if !scanner.Scan() {
			return fmt.Errorf("failed to read input")
		}
		version = strings.TrimSpace(scanner.Text())
	} else {
		fmt.Printf("Version: %s\n", version)
	}

	// Get description
	if description == "" {
		fmt.Print("Description (optional, press Enter to skip): ")
		if !scanner.Scan() {
			return fmt.Errorf("failed to read input")
		}
		description = strings.TrimSpace(scanner.Text())
	} else {
		fmt.Printf("Description: %s\n", description)
	}

	return runPackageAdd(packageName, provider, profile, version, description, packageID)
}

// selectProviderUI allows selection of a provider with UI
func selectProviderUI() (string, error) {
	// Load global config to check for default provider
	appConfig, err := config.LoadAppConfig()
	if err != nil {
		return "", fmt.Errorf("error loading app config: %w", err)
	}

	providersConfig, err := config.LoadProvidersConfig()
	if err != nil {
		return "", fmt.Errorf("error loading providers config: %w", err)
	}

	if len(providersConfig.Providers) == 0 {
		return "", fmt.Errorf("no providers available. Please add a provider first using 'al provider add'")
	}

	model := ui.NewProviderSelectModel(providersConfig.Providers, "Provider", appConfig.DefaultProvider)
	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		return "", fmt.Errorf("error running TUI: %w", err)
	}

	selected := model.GetSelected()
	if selected == "" {
		return "", fmt.Errorf("provider selection is required")
	}

	return selected, nil
}

// selectProfileUI allows selection of a profile with UI
func selectProfileUI() (string, error) {
	// Load global config to check for default profile
	appConfig, err := config.LoadAppConfig()
	if err != nil {
		return "", fmt.Errorf("error loading app config: %w", err)
	}

	profilesConfig, err := config.LoadProfilesConfig()
	if err != nil {
		return "", fmt.Errorf("error loading profiles config: %w", err)
	}

	if len(profilesConfig.Profiles) == 0 {
		return "", fmt.Errorf("no profiles available. Please add a profile first using 'al profile add'")
	}

	model := ui.NewProfileSelectModel(profilesConfig.Profiles, "Profile", appConfig.DefaultProfile)
	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		return "", fmt.Errorf("error running TUI: %w", err)
	}

	selected := model.GetSelected()
	if selected == "" {
		return "", fmt.Errorf("profile selection is required")
	}

	return selected, nil
}
