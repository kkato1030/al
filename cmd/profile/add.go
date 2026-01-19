package profile

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kkato1030/al/internal/config"
	"github.com/spf13/cobra"
)

// NewProfileAddCmd creates the profile add command
func NewProfileAddCmd() *cobra.Command {
	var description string
	var extends string
	var promoteTo string
	var packageDuplication string

	cmd := &cobra.Command{
		Use:   "add [profile-name]",
		Short: "Add a profile",
		Long:  "Add a new profile configuration. If profile-name is not provided, interactive mode will be used.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var name string
			if len(args) > 0 {
				name = args[0]
			}

			// If no arguments provided or flags are not set, use interactive mode
			if len(args) == 0 || (description == "" && extends == "" && promoteTo == "" && packageDuplication == "") {
				return runProfileAddInteractive(name, description, extends, promoteTo, packageDuplication)
			}

			return runProfileAdd(name, description, extends, promoteTo, packageDuplication)
		},
	}

	cmd.Flags().StringVarP(&description, "description", "d", "", "Description of the profile")
	cmd.Flags().StringVarP(&extends, "extends", "e", "", "Comma-separated list of profile names to extend")
	cmd.Flags().StringVarP(&promoteTo, "promote-to", "p", "", "Target location for promotion")
	cmd.Flags().StringVar(&packageDuplication, "package-duplication", "", "Package duplication policy: forbid, allow, or warn (default: warn)")

	return cmd
}

func runProfileAdd(name, description, extendsStr, promoteTo, packageDuplication string) error {
	// Parse extends if provided
	var extends []string
	if extendsStr != "" {
		extends = strings.Split(extendsStr, ",")
		// Trim whitespace from each profile name
		for i, e := range extends {
			extends[i] = strings.TrimSpace(e)
		}
	}

	// Validate that extended profiles exist
	if len(extends) > 0 {
		profilesConfig, err := config.LoadProfilesConfig()
		if err != nil {
			return fmt.Errorf("error loading profiles config: %w", err)
		}

		for _, extendName := range extends {
			found := false
			for _, p := range profilesConfig.Profiles {
				if p.Name == extendName {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("profile '%s' specified in extends does not exist", extendName)
			}
		}
	}

	// Validate package_duplication value
	if packageDuplication != "" {
		validValues := map[string]bool{"forbid": true, "allow": true, "warn": true}
		if !validValues[packageDuplication] {
			return fmt.Errorf("invalid package_duplication value: %s (must be forbid, allow, or warn)", packageDuplication)
		}
	} else {
		// Set default value
		packageDuplication = "warn"
	}

	profile := config.ProfileConfig{
		Name:               name,
		Description:        description,
		Extends:            extends,
		PromoteTo:          promoteTo,
		PackageDuplication: packageDuplication,
	}

	if err := config.AddOrUpdateProfile(profile); err != nil {
		return fmt.Errorf("error saving profile: %w", err)
	}

	fmt.Printf("Profile '%s' has been successfully added\n", name)
	return nil
}

func runProfileAddInteractive(name, description, extends, promoteTo, packageDuplication string) error {
	scanner := bufio.NewScanner(os.Stdin)

	// Get profile name
	if name == "" {
		fmt.Print("Profile name: ")
		if !scanner.Scan() {
			return fmt.Errorf("failed to read input")
		}
		name = strings.TrimSpace(scanner.Text())
		if name == "" {
			return fmt.Errorf("profile name is required")
		}
	} else {
		fmt.Printf("Profile name: %s\n", name)
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

	// Get extends (multiple selection)
	if extends == "" {
		selectedExtends, err := selectProfilesMultipleUI("Extends", name)
		if err != nil {
			return err
		}
		extends = strings.Join(selectedExtends, ",")
	} else {
		fmt.Printf("Extends: %s\n", extends)
	}

	// Get promote_to (single selection)
	if promoteTo == "" {
		selectedPromoteTo, err := selectProfileSingleUI("Promote to", name)
		if err != nil {
			return err
		}
		promoteTo = selectedPromoteTo
	} else {
		fmt.Printf("Promote to: %s\n", promoteTo)
	}

	// Get package_duplication
	if packageDuplication == "" {
		selectedPackageDuplication, err := selectPackageDuplicationUI()
		if err != nil {
			return err
		}
		packageDuplication = selectedPackageDuplication
	} else {
		fmt.Printf("Package duplication: %s\n", packageDuplication)
	}

	return runProfileAdd(name, description, extends, promoteTo, packageDuplication)
}

// selectProfilesMultipleUI allows multiple selection of profiles with UI
func selectProfilesMultipleUI(prompt string, excludeName string) ([]string, error) {
	profilesConfig, err := config.LoadProfilesConfig()
	if err != nil {
		return nil, fmt.Errorf("error loading profiles config: %w", err)
	}

	// Filter out the current profile being added
	availableProfiles := []config.ProfileConfig{}
	for _, p := range profilesConfig.Profiles {
		if p.Name != excludeName {
			availableProfiles = append(availableProfiles, p)
		}
	}

	if len(availableProfiles) == 0 {
		return []string{}, nil
	}

	model := NewOrderedMultiSelectModel(availableProfiles, prompt, excludeName)
	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		return nil, fmt.Errorf("error running TUI: %w", err)
	}

	return model.GetSelected(), nil
}

// selectProfileSingleUI allows single selection of a profile with UI
func selectProfileSingleUI(prompt string, excludeName string) (string, error) {
	profilesConfig, err := config.LoadProfilesConfig()
	if err != nil {
		return "", fmt.Errorf("error loading profiles config: %w", err)
	}

	// Filter out the current profile being added
	availableProfiles := []config.ProfileConfig{}
	for _, p := range profilesConfig.Profiles {
		if p.Name != excludeName {
			availableProfiles = append(availableProfiles, p)
		}
	}

	if len(availableProfiles) == 0 {
		return "", nil
	}

	model := NewSingleSelectModel(availableProfiles, prompt, excludeName)
	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		return "", fmt.Errorf("error running TUI: %w", err)
	}

	return model.GetSelected(), nil
}

// selectPackageDuplicationUI allows selection of package duplication policy with UI
func selectPackageDuplicationUI() (string, error) {
	model := NewPackageDuplicationSelectModel()
	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		return "", fmt.Errorf("error running TUI: %w", err)
	}

	return model.GetSelected(), nil
}

