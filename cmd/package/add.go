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

// NewPackageAddCmd creates the package add command
func NewPackageAddCmd() *cobra.Command {
	var provider string
	var profile string
	var version string
	var description string

	cmd := &cobra.Command{
		Use:   "add [package-name]",
		Short: "Add a package",
		Long:  "Add a package to a profile with a provider. If package-name is not provided, interactive mode will be used.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// If no package name provided, use fully interactive mode
			if len(args) == 0 {
				return runPackageAddInteractive("", provider, profile, version, description)
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

			return runPackageAdd(packageName, finalProvider, finalProfile, version, description)
		},
	}

	cmd.Flags().StringVarP(&provider, "provider", "p", "", "Provider name (required)")
	cmd.Flags().StringVarP(&profile, "profile", "f", "", "Profile name (required)")
	cmd.Flags().StringVarP(&version, "version", "v", "", "Package version (optional)")
	cmd.Flags().StringVarP(&description, "description", "d", "", "Package description (optional)")

	return cmd
}

func runPackageAdd(packageName, providerName, profile, version, description string) error {
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

	// Check if package already exists in config
	packagesConfig, err := config.LoadPackagesConfig()
	if err != nil {
		return fmt.Errorf("error loading packages config: %w", err)
	}

	packageExists := false
	for _, existingPkg := range packagesConfig.Packages {
		if existingPkg.Name == packageName && existingPkg.Provider == providerName && existingPkg.Profile == profile {
			packageExists = true
			break
		}
	}

	// Get provider instance
	var p provider.Provider
	switch providerName {
	case "brew":
		p = provider.NewBrewProvider()
	default:
		return fmt.Errorf("unsupported provider: %s", providerName)
	}

	// Install the package only if it doesn't exist in config
	if !packageExists {
		if err := p.InstallPackage(packageName); err != nil {
			return fmt.Errorf("error installing package: %w", err)
		}
	} else {
		fmt.Printf("Package '%s' already exists in config, skipping installation\n", packageName)
	}

	// Create package config
	pkg := config.PackageConfig{
		Name:        packageName,
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
		fmt.Printf("Package '%s' has been successfully updated in profile '%s' with provider '%s'\n", packageName, profile, providerName)
	} else {
		fmt.Printf("Package '%s' has been successfully added to profile '%s' with provider '%s'\n", packageName, profile, providerName)
	}
	return nil
}

func runPackageAddInteractive(packageName, provider, profile, version, description string) error {
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
		selectedProvider, err := selectProvider(scanner)
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
		selectedProfile, err := selectProfile(scanner)
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

	return runPackageAdd(packageName, provider, profile, version, description)
}

// selectProvider allows selection of a provider
func selectProvider(scanner *bufio.Scanner) (string, error) {
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

	// Find default provider index
	defaultIdx := -1
	if appConfig.DefaultProvider != "" {
		for i, p := range providersConfig.Providers {
			if p.Name == appConfig.DefaultProvider {
				defaultIdx = i
				break
			}
		}
	}

	fmt.Printf("\nProviders:\n")
	for i, p := range providersConfig.Providers {
		fmt.Printf("  %d. %s", i+1, p.Name)
		if i == defaultIdx {
			fmt.Printf(" (default)")
		}
		if p.Version != "" {
			fmt.Printf(" (version: %s)", p.Version)
		}
		fmt.Println()
	}

	// Build prompt with default info
	prompt := "Select provider (number"
	if defaultIdx >= 0 {
		prompt += fmt.Sprintf(", default: %s, press Enter to use", providersConfig.Providers[defaultIdx].Name)
	}
	prompt += "): "
	fmt.Print(prompt)

	if !scanner.Scan() {
		return "", fmt.Errorf("failed to read input")
	}

	input := strings.TrimSpace(scanner.Text())
	if input == "" {
		if defaultIdx >= 0 {
			return providersConfig.Providers[defaultIdx].Name, nil
		}
		return "", fmt.Errorf("provider selection is required")
	}

	idx, err := strconv.Atoi(input)
	if err != nil {
		return "", fmt.Errorf("invalid number: %s", input)
	}

	if idx < 1 || idx > len(providersConfig.Providers) {
		return "", fmt.Errorf("number %d is out of range (1-%d)", idx, len(providersConfig.Providers))
	}

	return providersConfig.Providers[idx-1].Name, nil
}

// selectProfile allows selection of a profile
func selectProfile(scanner *bufio.Scanner) (string, error) {
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

	// Find default profile index
	defaultIdx := -1
	if appConfig.DefaultProfile != "" {
		for i, p := range profilesConfig.Profiles {
			if p.Name == appConfig.DefaultProfile {
				defaultIdx = i
				break
			}
		}
	}

	fmt.Printf("\nProfiles:\n")
	for i, p := range profilesConfig.Profiles {
		fmt.Printf("  %d. %s", i+1, p.Name)
		if i == defaultIdx {
			fmt.Printf(" (default)")
		}
		if p.Description != "" {
			fmt.Printf(" - %s", p.Description)
		}
		fmt.Println()
	}

	// Build prompt with default info
	prompt := "Select profile (number"
	if defaultIdx >= 0 {
		prompt += fmt.Sprintf(", default: %s, press Enter to use", profilesConfig.Profiles[defaultIdx].Name)
	}
	prompt += "): "
	fmt.Print(prompt)

	if !scanner.Scan() {
		return "", fmt.Errorf("failed to read input")
	}

	input := strings.TrimSpace(scanner.Text())
	if input == "" {
		if defaultIdx >= 0 {
			return profilesConfig.Profiles[defaultIdx].Name, nil
		}
		return "", fmt.Errorf("profile selection is required")
	}

	idx, err := strconv.Atoi(input)
	if err != nil {
		return "", fmt.Errorf("invalid number: %s", input)
	}

	if idx < 1 || idx > len(profilesConfig.Profiles) {
		return "", fmt.Errorf("number %d is out of range (1-%d)", idx, len(profilesConfig.Profiles))
	}

	return profilesConfig.Profiles[idx-1].Name, nil
}
