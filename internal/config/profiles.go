package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ProfileConfig represents a profile configuration
type ProfileConfig struct {
	Name               string   `json:"name"`
	Description        string   `json:"description,omitempty"`
	Stage              string   `json:"stage,omitempty"` // "stable" or "trial"
	Extends            []string `json:"extends,omitempty"`
	PromoteTo          string   `json:"promote_to,omitempty"`
	PackageDuplication string   `json:"package_duplication,omitempty"`
	AutoSync           *bool    `json:"auto_sync,omitempty"` // nil/true = include in "al sync" (default), false = exclude unless --profile
	// ReviewDays is the number of days after which packages must be reviewed. nil or 0 = not a review target. Set to e.g. 30 to require review every 30 days (based on package reviewed_at).
	ReviewDays *int `json:"review_days,omitempty"`
}

// ProfilesConfig represents the collection of profile configurations
type ProfilesConfig struct {
	Profiles []ProfileConfig `json:"profiles"`
}

// LoadProfilesConfig loads the profiles configuration from JSON file
func LoadProfilesConfig() (*ProfilesConfig, error) {
	configPath, err := GetProfilesConfigPath()
	if err != nil {
		return nil, err
	}

	// If file doesn't exist, return empty config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &ProfilesConfig{Profiles: []ProfileConfig{}}, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config ProfilesConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// SaveProfilesConfig saves the profiles configuration to JSON file
func SaveProfilesConfig(config *ProfilesConfig) error {
	// Ensure config directory exists
	if err := EnsureConfigDir(); err != nil {
		return err
	}

	configPath, err := GetProfilesConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

// AddOrUpdateProfile adds or updates a profile in the configuration
func AddOrUpdateProfile(profile ProfileConfig) error {
	config, err := LoadProfilesConfig()
	if err != nil {
		return err
	}

	// Check if profile already exists
	found := false
	for i, p := range config.Profiles {
		if p.Name == profile.Name {
			config.Profiles[i] = profile
			found = true
			break
		}
	}

	// If not found, add it
	if !found {
		config.Profiles = append(config.Profiles, profile)
	}

	return SaveProfilesConfig(config)
}

// GetProfile returns a profile configuration by name
func GetProfile(name string) (*ProfileConfig, error) {
	config, err := LoadProfilesConfig()
	if err != nil {
		return nil, err
	}

	for _, p := range config.Profiles {
		if p.Name == name {
			return &p, nil
		}
	}

	return nil, nil // Profile not found
}

// RemoveProfile removes a profile from the configuration
// TODO: In the future, this should also remove packages associated with the profile
// and any generated config files related to the profile
func RemoveProfile(name string) error {
	config, err := LoadProfilesConfig()
	if err != nil {
		return err
	}

	// Find and remove the profile
	found := false
	for i, p := range config.Profiles {
		if p.Name == name {
			// Remove the profile by creating a new slice without it
			config.Profiles = append(config.Profiles[:i], config.Profiles[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		return nil // Profile not found, but don't return an error
	}

	return SaveProfilesConfig(config)
}

// GetProfilesConfigPath returns the path to the profiles.json file
func GetProfilesConfigPath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(configDir, "profiles.json"), nil
}

// ValidateProfileName validates that a profile name contains only allowed characters
// Allowed characters: -, _, #, @, ., alphanumeric
func ValidateProfileName(name string) error {
	if name == "" {
		return fmt.Errorf("profile name cannot be empty")
	}

	// Regular expression to match allowed characters: -, _, #, @, ., and alphanumeric
	validPattern := regexp.MustCompile(`^[a-zA-Z0-9_#@.-]+$`)
	if !validPattern.MatchString(name) {
		return fmt.Errorf("profile name '%s' contains invalid characters. Only alphanumeric characters, -, _, #, @, and . are allowed", name)
	}

	return nil
}

// ParseProfileName parses a profile name into profile_name and stage_name
// Format: profile_name.stage_name
// Returns profile_name and stage_name (empty string if no stage)
func ParseProfileName(fullName string) (profileName string, stageName string, err error) {
	if err := ValidateProfileName(fullName); err != nil {
		return "", "", err
	}

	parts := strings.SplitN(fullName, ".", 2)
	if len(parts) == 1 {
		// No stage specified
		return parts[0], "", nil
	}

	// Both profile_name and stage_name must be validated
	if err := ValidateProfileName(parts[0]); err != nil {
		return "", "", fmt.Errorf("invalid profile_name in '%s': %w", fullName, err)
	}
	if err := ValidateProfileName(parts[1]); err != nil {
		return "", "", fmt.Errorf("invalid stage_name in '%s': %w", fullName, err)
	}

	return parts[0], parts[1], nil
}

// BuildProfileName builds a full profile name from profile_name and stage_name
// Format: profile_name.stage_name (or just profile_name if stage is empty)
func BuildProfileName(profileName, stageName string) (string, error) {
	if err := ValidateProfileName(profileName); err != nil {
		return "", err
	}

	if stageName == "" {
		return profileName, nil
	}

	if err := ValidateProfileName(stageName); err != nil {
		return "", err
	}

	return fmt.Sprintf("%s.%s", profileName, stageName), nil
}

// IsAutoSyncEnabled returns true if the profile should be included in "al sync" by default (or with --all).
// nil or true means include, false means exclude unless explicitly specified with --profile.
func IsAutoSyncEnabled(p ProfileConfig) bool {
	return p.AutoSync == nil || *p.AutoSync
}

// profileByName returns a profile from the config by name, or nil if not found.
func profileByName(cfg *ProfilesConfig, name string) *ProfileConfig {
	for i := range cfg.Profiles {
		if cfg.Profiles[i].Name == name {
			return &cfg.Profiles[i]
		}
	}
	return nil
}

// ResolveProfileWithExtends returns the given profile name and all profiles it extends (recursively), with no duplicates.
func ResolveProfileWithExtends(name string) ([]string, error) {
	cfg, err := LoadProfilesConfig()
	if err != nil {
		return nil, err
	}
	seen := make(map[string]bool)
	var out []string
	var visit func(string) error
	visit = func(n string) error {
		if seen[n] {
			return nil
		}
		seen[n] = true
		out = append(out, n)
		p := profileByName(cfg, n)
		if p == nil {
			return fmt.Errorf("profile not found: %s", n)
		}
		for _, e := range p.Extends {
			if err := visit(e); err != nil {
				return err
			}
		}
		return nil
	}
	if err := visit(name); err != nil {
		return nil, err
	}
	return out, nil
}

// GetReviewDays returns the review period in days for the profile, and whether review is required.
// No review_days (nil or <= 0) → not a review target. Has review_days → packages use reviewed_at + review_days for due date.
func GetReviewDays(profileName string) (days int, hasReview bool, err error) {
	p, err := GetProfile(profileName)
	if err != nil || p == nil {
		return 0, false, err
	}
	if p.ReviewDays == nil || *p.ReviewDays <= 0 {
		return 0, false, nil
	}
	return *p.ReviewDays, true, nil
}

// GetSyncTargetProfileNames returns profile names to sync based on mode.
// Mode: "default" = default_profile + extends, filtered by AutoSync; "all" = all profiles with AutoSync; "profile" = given name + extends (no filter).
func GetSyncTargetProfileNames(mode string, profileName string) ([]string, error) {
	profilesCfg, err := LoadProfilesConfig()
	if err != nil {
		return nil, err
	}
	switch mode {
	case "profile":
		if profileName == "" {
			return nil, fmt.Errorf("profile name required for mode profile")
		}
		return ResolveProfileWithExtends(profileName)
	case "all":
		var names []string
		for _, p := range profilesCfg.Profiles {
			if IsAutoSyncEnabled(p) {
				names = append(names, p.Name)
			}
		}
		return names, nil
	case "default":
		appCfg, err := LoadAppConfig()
		if err != nil {
			return nil, err
		}
		def := strings.TrimSpace(appCfg.DefaultProfile)
		if def == "" {
			return []string{}, nil
		}
		all, err := ResolveProfileWithExtends(def)
		if err != nil {
			return nil, err
		}
		var names []string
		for _, n := range all {
			p := profileByName(profilesCfg, n)
			if p != nil && IsAutoSyncEnabled(*p) {
				names = append(names, n)
			}
		}
		return names, nil
	default:
		return nil, fmt.Errorf("invalid sync mode: %s", mode)
	}
}

// SortProfilesForDisplay sorts profiles for display in the following order:
// 1. Group by base profile name (part before the dot)
// 2. Within each group: stable (no stage or explicit "stable") before trial
// 3. Groups ordered: "core" first, then alphabetically
func SortProfilesForDisplay(profiles []ProfileConfig) []ProfileConfig {
	if len(profiles) == 0 {
		return profiles
	}

	// Group profiles by base name
	type profileGroup struct {
		baseName string
		profiles []ProfileConfig
	}

	groupMap := make(map[string]*profileGroup)
	var groupOrder []string

	for _, p := range profiles {
		baseName, _, err := ParseProfileName(p.Name)
		if err != nil {
			// If parsing fails, use the full name as base
			baseName = p.Name
		}

		if _, exists := groupMap[baseName]; !exists {
			groupMap[baseName] = &profileGroup{baseName: baseName}
			groupOrder = append(groupOrder, baseName)
		}

		// Store both baseName and stageName for sorting within group
		groupMap[baseName].profiles = append(groupMap[baseName].profiles, p)
	}

	// Sort groups: "core" first, then alphabetically
	sortGroupOrder := func(groups []string) []string {
		var coreGroups []string
		var otherGroups []string

		for _, g := range groups {
			if g == "core" {
				coreGroups = append(coreGroups, g)
			} else {
				otherGroups = append(otherGroups, g)
			}
		}

		// Sort other groups alphabetically
		for i := 0; i < len(otherGroups); i++ {
			for j := i + 1; j < len(otherGroups); j++ {
				if strings.ToLower(otherGroups[i]) > strings.ToLower(otherGroups[j]) {
					otherGroups[i], otherGroups[j] = otherGroups[j], otherGroups[i]
				}
			}
		}

		return append(coreGroups, otherGroups...)
	}

	groupOrder = sortGroupOrder(groupOrder)

	// Sort profiles within each group: stable before trial
	for _, group := range groupMap {
		profiles := group.profiles
		for i := 0; i < len(profiles); i++ {
			for j := i + 1; j < len(profiles); j++ {
				_, stageI, _ := ParseProfileName(profiles[i].Name)
				_, stageJ, _ := ParseProfileName(profiles[j].Name)

				// Stable (empty or "stable") comes before trial
				isStableI := stageI == "" || stageI == "stable"
				isStableJ := stageJ == "" || stageJ == "stable"

				if !isStableI && isStableJ {
					profiles[i], profiles[j] = profiles[j], profiles[i]
				} else if isStableI == isStableJ {
					// If both are stable or both are trial, sort alphabetically by stage name
					if strings.ToLower(stageI) > strings.ToLower(stageJ) {
						profiles[i], profiles[j] = profiles[j], profiles[i]
					}
				}
			}
		}
		group.profiles = profiles
	}

	// Build the final sorted list
	var result []ProfileConfig
	for _, baseName := range groupOrder {
		result = append(result, groupMap[baseName].profiles...)
	}

	return result
}
