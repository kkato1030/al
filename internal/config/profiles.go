package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// ProfileConfig represents a profile configuration
type ProfileConfig struct {
	Name               string   `json:"name"`
	Description        string   `json:"description,omitempty"`
	Extends            []string `json:"extends,omitempty"`
	PromoteTo          string   `json:"promote_to,omitempty"`
	PackageDuplication string   `json:"package_duplication,omitempty"`
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

// GetProfilesConfigPath returns the path to the profiles.json file
func GetProfilesConfigPath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(configDir, "profiles.json"), nil
}
