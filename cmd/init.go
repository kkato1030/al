package cmd

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kkato1030/al/internal/config"
	"github.com/kkato1030/al/internal/output"
	"github.com/kkato1030/al/internal/provider"
	"github.com/kkato1030/al/internal/ui"
	"github.com/spf13/cobra"
)

// NewInitCmd creates the init command
func NewInitCmd() *cobra.Command {
	var guided bool

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize al with standard configuration",
		Long:  "Initialize al for first-time users. Sets up stable-trial profiles (core and core.trial), provider (brew), and default settings. Use --guided for interactive setup.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if guided {
				return runInitGuided()
			}
			return runInit(cmd, args)
		},
	}

	cmd.Flags().BoolVar(&guided, "guided", false, "Use interactive guided setup")

	return cmd
}

func runInit(cmd *cobra.Command, args []string) error {
	// Ensure config directory exists
	if err := config.EnsureConfigDir(); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Create .gitignore for logs directory
	if err := config.EnsureGitignore(); err != nil {
		output.Warning("Failed to create .gitignore: %v", err)
		// Don't fail init if .gitignore creation fails
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
		output.Success("Profile %s set up", profile.Name)
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
	output.Success("Provider brew set up")

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

	output.Success("Initialization complete")
	output.Info("You can start using al with 'al package add'.")
	return nil
}

func runInitGuided() error {
	fmt.Println("Welcome to al! Let's set up your environment.")

	// Ensure config directory exists
	if err := config.EnsureConfigDir(); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Create .gitignore for logs directory
	if err := config.EnsureGitignore(); err != nil {
		output.Warning("Failed to create .gitignore: %v", err)
	}

	// Prompt 1: Profile setup strategy
	setupModel := ui.NewSelectModel(
		"How would you like to set up profiles?",
		[]string{"Single profile (core only)", "Multiple profiles (core + additional)"},
	)
	p := tea.NewProgram(setupModel)
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running UI: %w", err)
	}
	setupChoice := setupModel.GetSelected()
	if setupChoice == "" {
		return fmt.Errorf("profile setup selection is required")
	}

	// Prompt 2: Additional profile names (if multiple profiles chosen)
	var additionalProfiles []string
	if setupChoice == "Multiple profiles (core + additional)" {
		nameModel := ui.NewTextInputModel("Enter additional profile names (comma-separated)", false)
		p = tea.NewProgram(nameModel)
		if _, err := p.Run(); err != nil {
			return fmt.Errorf("error running UI: %w", err)
		}
		profileInput := nameModel.GetValue()
		if profileInput != "" {
			// Parse comma-separated profile names
			for _, name := range strings.Split(profileInput, ",") {
				trimmed := strings.TrimSpace(name)
				if trimmed != "" && trimmed != "core" {
					additionalProfiles = append(additionalProfiles, trimmed)
				}
			}
		}
	}

	// Prompt 3: Enable trial workflow
	trialModel := ui.NewSelectModel(
		"Enable trial workflow for experimental packages?",
		[]string{"yes", "no"},
	)
	p = tea.NewProgram(trialModel)
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running UI: %w", err)
	}
	enableTrial := trialModel.GetSelected() == "yes"

	// Prompt 4: Review period (if trial enabled)
	reviewDays := 1 // default
	if enableTrial {
		reviewModel := ui.NewSelectModel(
			"Review period for trial packages?",
			[]string{"1 day", "7 days", "14 days", "30 days", "60 days"},
		)
		p = tea.NewProgram(reviewModel)
		if _, err := p.Run(); err != nil {
			return fmt.Errorf("error running UI: %w", err)
		}
		reviewSelection := reviewModel.GetSelected()
		switch reviewSelection {
		case "1 day":
			reviewDays = 1
		case "7 days":
			reviewDays = 7
		case "14 days":
			reviewDays = 14
		case "30 days":
			reviewDays = 30
		case "60 days":
			reviewDays = 60
		}
	}

	// Apply template based on trial choice for core profile
	var templateName string
	if enableTrial {
		templateName = "stable-trial"
	} else {
		templateName = "stable-only"
	}

	template, err := config.GetTemplate(templateName)
	if err != nil {
		return fmt.Errorf("failed to get template '%s': %w", templateName, err)
	}

	// Create core profile(s)
	coreProfiles, err := config.ApplyTemplate(template, "core")
	if err != nil {
		return fmt.Errorf("failed to apply template: %w", err)
	}

	// Save core profiles with custom review days if trial enabled
	for _, profile := range coreProfiles {
		// Set default package_duplication if not set
		if profile.PackageDuplication == "" {
			profile.PackageDuplication = "warn"
		}

		// Set custom review days for trial profile
		if enableTrial && profile.Stage == "trial" {
			profile.ReviewDays = &reviewDays
		}

		if err := config.AddOrUpdateProfile(profile); err != nil {
			return fmt.Errorf("failed to add profile '%s': %w", profile.Name, err)
		}
		output.Success("Profile %s set up", profile.Name)
	}

	// Create additional profiles if any
	for _, profileName := range additionalProfiles {
		additionalProfileConfigs, err := config.ApplyTemplate(template, profileName)
		if err != nil {
			return fmt.Errorf("failed to apply template for '%s': %w", profileName, err)
		}

		for _, profile := range additionalProfileConfigs {
			// Set default package_duplication if not set
			if profile.PackageDuplication == "" {
				profile.PackageDuplication = "warn"
			}

			// Set custom review days for trial profile
			if enableTrial && profile.Stage == "trial" {
				profile.ReviewDays = &reviewDays
			}

			if err := config.AddOrUpdateProfile(profile); err != nil {
				return fmt.Errorf("failed to add profile '%s': %w", profile.Name, err)
			}
			output.Success("Profile %s set up", profile.Name)
		}
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
	output.Success("Provider brew set up")

	// Set default_profile, default_provider, default_stage
	var defaultProfile string
	var defaultStage string
	if enableTrial {
		defaultProfile = "core.trial"
		defaultStage = "trial"
	} else {
		defaultProfile = "core"
		defaultStage = "stable"
	}

	appConfig := &config.AppConfig{
		DefaultProfile:  defaultProfile,
		DefaultProvider: "brew",
		DefaultStage:    defaultStage,
	}
	if err := config.SaveAppConfig(appConfig); err != nil {
		return fmt.Errorf("failed to save app config: %w", err)
	}
	fmt.Printf("Default settings: profile=%s, provider=brew, stage=%s\n", defaultProfile, defaultStage)

	output.Success("Initialization complete")
	fmt.Printf("\nConfiguration summary:\n")
	fmt.Printf("  Profile setup: %s\n", setupChoice)
	if len(additionalProfiles) > 0 {
		fmt.Printf("  Additional profiles: %s\n", strings.Join(additionalProfiles, ", "))
	}
	trialStatus := "disabled"
	if enableTrial {
		trialStatus = "enabled"
	}
	fmt.Printf("  Trial workflow: %s\n", trialStatus)
	if enableTrial {
		fmt.Printf("  Review period: %d days\n", reviewDays)
	}
	fmt.Printf("\nYou can start using al with 'al package add'.\n")

	return nil
}
