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
