package profile

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kkato1030/al/internal/config"
	"github.com/kkato1030/al/internal/ui"
	"github.com/spf13/cobra"
)

// NewProfileAddCmd creates the profile add command
func NewProfileAddCmd() *cobra.Command {
	var description string
	var extends string
	var promoteTo string
	var packageDuplication string
	var templateName string

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

			// If template is specified, use template mode
			if templateName != "" {
				return runProfileAddFromTemplate(name, templateName, description)
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
	cmd.Flags().StringVarP(&templateName, "template", "t", "", "Template name to use for creating profiles")

	return cmd
}

func runProfileAdd(name, description, extendsStr, promoteTo, packageDuplication string) error {
	// Validate profile name
	if err := config.ValidateProfileName(name); err != nil {
		return fmt.Errorf("invalid profile name: %w", err)
	}

	// Parse extends if provided
	var extends []string
	if extendsStr != "" {
		extends = strings.Split(extendsStr, ",")
		// Trim whitespace from each profile name
		for i, e := range extends {
			extends[i] = strings.TrimSpace(e)
			// Validate each extended profile name
			if err := config.ValidateProfileName(extends[i]); err != nil {
				return fmt.Errorf("invalid extended profile name '%s': %w", extends[i], err)
			}
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

	// Validate promoteTo if provided
	if promoteTo != "" {
		if err := config.ValidateProfileName(promoteTo); err != nil {
			return fmt.Errorf("invalid promote_to profile name '%s': %w", promoteTo, err)
		}
	}

	profile := config.ProfileConfig{
		Name:               name,
		Description:        description,
		Extends:            extends,
		PromoteTo:          promoteTo,
		PackageDuplication: packageDuplication,
	}

	// Validate stage if provided
	if profile.Stage != "" {
		if err := validateStage(profile.Stage); err != nil {
			return err
		}
	}

	if err := config.AddOrUpdateProfile(profile); err != nil {
		return fmt.Errorf("error saving profile: %w", err)
	}

	fmt.Printf("Profile '%s' has been successfully added\n", name)
	return nil
}

// runProfileAddFromTemplate creates profiles from a template
func runProfileAddFromTemplate(profileName, templateName, description string) error {
	// Get template
	template, err := config.GetTemplate(templateName)
	if err != nil {
		return fmt.Errorf("error getting template: %w", err)
	}

	// Get profile name if not provided
	if profileName == "" {
		scanner := bufio.NewScanner(os.Stdin)
		fmt.Print("Profile name: ")
		if !scanner.Scan() {
			return fmt.Errorf("failed to read input")
		}
		profileName = strings.TrimSpace(scanner.Text())
		if profileName == "" {
			return fmt.Errorf("profile name is required")
		}
	}

	// Apply template
	profiles, err := config.ApplyTemplate(template, profileName)
	if err != nil {
		return fmt.Errorf("error applying template: %w", err)
	}

	// Sort profiles by dependencies: profiles that are extended by others should be created first
	profiles = sortProfilesByDependencies(profiles)

	// Validate and save each profile
	for _, profile := range profiles {
		// Validate profile name
		if err := config.ValidateProfileName(profile.Name); err != nil {
			return fmt.Errorf("invalid profile name '%s': %w", profile.Name, err)
		}

		// Set description if provided
		if description != "" {
			profile.Description = description
		}

		// Validate stage
		if profile.Stage != "" {
			if err := validateStage(profile.Stage); err != nil {
				return fmt.Errorf("invalid stage in template: %w", err)
			}
		}

		// Validate extends profile names
		for _, ext := range profile.Extends {
			if err := config.ValidateProfileName(ext); err != nil {
				return fmt.Errorf("invalid extended profile name '%s': %w", ext, err)
			}
		}

		// Validate promoteTo if provided
		if profile.PromoteTo != "" {
			if err := config.ValidateProfileName(profile.PromoteTo); err != nil {
				return fmt.Errorf("invalid promote_to profile name '%s': %w", profile.PromoteTo, err)
			}
		}

		// Validate that extended profiles exist (if any)
		// Note: We validate after sorting, so dependencies should be created first
		if len(profile.Extends) > 0 {
			profilesConfig, err := config.LoadProfilesConfig()
			if err != nil {
				return fmt.Errorf("error loading profiles config: %w", err)
			}

			for _, extendName := range profile.Extends {
				// Check if the extended profile is in the profiles we're about to create
				foundInNew := false
				for _, newProfile := range profiles {
					if newProfile.Name == extendName {
						foundInNew = true
						break
					}
				}

				if !foundInNew {
					// Check if it exists in existing profiles
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
		}

		// Set default package_duplication if not set
		if profile.PackageDuplication == "" {
			profile.PackageDuplication = "warn"
		}

		// Save profile
		if err := config.AddOrUpdateProfile(profile); err != nil {
			return fmt.Errorf("error saving profile '%s': %w", profile.Name, err)
		}

		fmt.Printf("Profile '%s' has been successfully added\n", profile.Name)
	}

	return nil
}

// validateStage validates that stage is either "stable" or "trial"
func validateStage(stage string) error {
	validStages := map[string]bool{"stable": true, "trial": true}
	if !validStages[stage] {
		return fmt.Errorf("invalid stage value: %s (must be stable or trial)", stage)
	}
	return nil
}

// sortProfilesByDependencies sorts profiles so that profiles that are extended by others come first
func sortProfilesByDependencies(profiles []config.ProfileConfig) []config.ProfileConfig {
	// Create a map of profile names to their indices
	nameToIndex := make(map[string]int)
	for i, p := range profiles {
		nameToIndex[p.Name] = i
	}

	// Create a map of dependencies (which profiles extend which)
	extendedBy := make(map[string][]int) // profile name -> indices of profiles that extend it
	for i, p := range profiles {
		for _, ext := range p.Extends {
			if _, exists := nameToIndex[ext]; exists {
				extendedBy[ext] = append(extendedBy[ext], i)
			}
		}
	}

	// Topological sort: profiles with no dependencies or dependencies that are not in the list come first
	sorted := make([]config.ProfileConfig, 0, len(profiles))
	added := make(map[int]bool)

	// Add profiles that don't extend anything in the list first
	for i, p := range profiles {
		hasInternalDeps := false
		for _, ext := range p.Extends {
			if _, exists := nameToIndex[ext]; exists {
				hasInternalDeps = true
				break
			}
		}
		if !hasInternalDeps {
			sorted = append(sorted, p)
			added[i] = true
		}
	}

	// Add remaining profiles (those that extend profiles in the list)
	// Simple approach: add them in order, but only if their dependencies are already added
	for len(sorted) < len(profiles) {
		progress := false
		for i, p := range profiles {
			if added[i] {
				continue
			}

			// Check if all dependencies are already added
			allDepsAdded := true
			for _, ext := range p.Extends {
				if depIdx, exists := nameToIndex[ext]; exists {
					if !added[depIdx] {
						allDepsAdded = false
						break
					}
				}
			}

			if allDepsAdded {
				sorted = append(sorted, p)
				added[i] = true
				progress = true
			}
		}

		// If no progress was made, there might be a circular dependency or missing dependency
		// In that case, just add the remaining profiles in order
		if !progress {
			for i, p := range profiles {
				if !added[i] {
					sorted = append(sorted, p)
					added[i] = true
				}
			}
			break
		}
	}

	return sorted
}

func runProfileAddInteractive(name, description, extends, promoteTo, packageDuplication string) error {
	// First, ask if user wants to use a template
	templates, err := config.GetAllTemplates()
	if err != nil {
		return fmt.Errorf("error loading templates: %w", err)
	}

	// If templates are available, ask user if they want to use one
	if len(templates) > 0 {
		selectedTemplate, err := selectTemplateUI("Use template? (press Enter to skip)", templates)
		if err != nil {
			return err
		}

		if selectedTemplate != "" {
			// Use template mode
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

			return runProfileAddFromTemplate(name, selectedTemplate, description)
		}
	}

	// Continue with manual profile creation
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

// selectTemplateUI allows selection of a template with UI
func selectTemplateUI(prompt string, templates []config.ProfileTemplate) (string, error) {
	if len(templates) == 0 {
		return "", nil
	}

	// Convert templates to ProfileConfig for UI compatibility
	items := make([]config.ProfileConfig, len(templates))
	for i, tmpl := range templates {
		items[i] = config.ProfileConfig{
			Name:        tmpl.Name,
			Description: fmt.Sprintf("Creates %d profile(s)", len(tmpl.Profiles)),
		}
	}

	model := ui.NewSingleSelectModel(items, prompt, "")
	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		return "", fmt.Errorf("error running TUI: %w", err)
	}

	return model.GetSelected(), nil
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

	model := ui.NewOrderedMultiSelectModel(availableProfiles, prompt, excludeName)
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

	model := ui.NewSingleSelectModel(availableProfiles, prompt, excludeName)
	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		return "", fmt.Errorf("error running TUI: %w", err)
	}

	return model.GetSelected(), nil
}

// selectPackageDuplicationUI allows selection of package duplication policy with UI
func selectPackageDuplicationUI() (string, error) {
	model := ui.NewPackageDuplicationSelectModel()
	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		return "", fmt.Errorf("error running TUI: %w", err)
	}

	return model.GetSelected(), nil
}

