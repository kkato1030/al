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
		Long:  "Initialize al for first-time users. Sets up standard profile (core with trial stage), provider (brew), and default settings.",
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

	// Add profile: core (stable-trial) -> profile "core" with stage "trial"
	coreProfile := config.ProfileConfig{
		Name:               "core",
		Stage:              "trial",
		PackageDuplication: "warn",
	}
	if err := config.AddOrUpdateProfile(coreProfile); err != nil {
		return fmt.Errorf("failed to add profile: %w", err)
	}
	fmt.Println("Profile 'core' (stage: trial) has been set up")

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
		DefaultProfile:  "core",
		DefaultProvider: "brew",
		DefaultStage:    "trial",
	}
	if err := config.SaveAppConfig(appConfig); err != nil {
		return fmt.Errorf("failed to save app config: %w", err)
	}
	fmt.Println("Default settings: profile=core, provider=brew, stage=trial")

	fmt.Println("\nInitialization complete. You can start using al with 'al package add'.")
	return nil
}
