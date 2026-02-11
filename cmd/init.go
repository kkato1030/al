package cmd

import (
	"fmt"
	"time"

	"github.com/kkato1030/al/internal/config"
	"github.com/kkato1030/al/internal/provider"
	"github.com/spf13/cobra"
)

// NewInitCmd creates the init command
func NewInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize al with standard configuration",
		Long:  "Initialize al for first-time users. Sets up stable-trial profiles (core and core.trial), provider (brew), and default settings.",
		Args:  cobra.NoArgs,
		RunE:  runInit,
	}
	return cmd
}

func runInit(cmd *cobra.Command, args []string) error {
	// Ensure config directory exists
	if err := config.EnsureConfigDir(); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Get stable-trial template
	template, err := config.GetTemplate("stable-trial")
	if err != nil {
		return fmt.Errorf("failed to get stable-trial template: %w", err)
	}

	// Apply template to create core and core.trial profiles
	profiles, err := config.ApplyTemplate(template, "core")
	if err != nil {
		return fmt.Errorf("failed to apply stable-trial template: %w", err)
	}

	// Save each profile
	for _, profile := range profiles {
		// Set default package_duplication if not set
		if profile.PackageDuplication == "" {
			profile.PackageDuplication = "warn"
		}

		if err := config.AddOrUpdateProfile(profile); err != nil {
			return fmt.Errorf("failed to add profile '%s': %w", profile.Name, err)
		}
		fmt.Printf("Profile '%s' has been set up\n", profile.Name)
	}

	// Add provider: brew
	brewProvider := provider.NewBrewProvider()
	version := ""
	if installed, err := brewProvider.CheckInstalled(); err == nil && installed {
		if v, err := brewProvider.GetVersion(); err == nil {
			version = v
		}
	}
	providerConfig := config.ProviderConfig{
		Name:        "brew",
		InstalledAt: time.Now(),
		Version:     version,
	}
	if err := config.AddOrUpdateProvider(providerConfig); err != nil {
		return fmt.Errorf("failed to add provider: %w", err)
	}
	fmt.Println("Provider 'brew' has been set up")

	// Set default_profile, default_provider, default_stage
	appConfig := &config.AppConfig{
		DefaultProfile:  "core.trial",
		DefaultProvider: "brew",
		DefaultStage:    "trial",
	}
	if err := config.SaveAppConfig(appConfig); err != nil {
		return fmt.Errorf("failed to save app config: %w", err)
	}
	fmt.Println("Default settings: profile=core.trial, provider=brew, stage=trial")

	fmt.Println("\nInitialization complete. You can start using al with 'al package add'.")
	return nil
}
