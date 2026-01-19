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
		Use:   "add <package-name>",
		Short: "Add a package",
		Long:  "Add a package to a profile with a provider. If required flags are not provided, interactive mode will be used.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			packageName := args[0]

			// If required flags are not set, use interactive mode
			if provider == "" || profile == "" {
				return runPackageAddInteractive(packageName, provider, profile, version, description)
			}

			return runPackageAdd(packageName, provider, profile, version, description)
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

	// Get package name
	fmt.Printf("Package name: %s\n", packageName)

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
	providersConfig, err := config.LoadProvidersConfig()
	if err != nil {
		return "", fmt.Errorf("error loading providers config: %w", err)
	}

	if len(providersConfig.Providers) == 0 {
		return "", fmt.Errorf("no providers available. Please add a provider first using 'al provider add'")
	}

	fmt.Printf("\nProviders:\n")
	for i, p := range providersConfig.Providers {
		fmt.Printf("  %d. %s", i+1, p.Name)
		if p.Version != "" {
			fmt.Printf(" (version: %s)", p.Version)
		}
		fmt.Println()
	}
	fmt.Print("Select provider (number): ")

	if !scanner.Scan() {
		return "", fmt.Errorf("failed to read input")
	}

	input := strings.TrimSpace(scanner.Text())
	if input == "" {
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
	profilesConfig, err := config.LoadProfilesConfig()
	if err != nil {
		return "", fmt.Errorf("error loading profiles config: %w", err)
	}

	if len(profilesConfig.Profiles) == 0 {
		return "", fmt.Errorf("no profiles available. Please add a profile first using 'al profile add'")
	}

	fmt.Printf("\nProfiles:\n")
	for i, p := range profilesConfig.Profiles {
		fmt.Printf("  %d. %s", i+1, p.Name)
		if p.Description != "" {
			fmt.Printf(" - %s", p.Description)
		}
		fmt.Println()
	}
	fmt.Print("Select profile (number): ")

	if !scanner.Scan() {
		return "", fmt.Errorf("failed to read input")
	}

	input := strings.TrimSpace(scanner.Text())
	if input == "" {
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
