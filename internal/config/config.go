package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// GetConfigDir returns the configuration directory path.
// It checks AL_HOME environment variable first, then defaults to ~/.al
func GetConfigDir() (string, error) {
	alHome := os.Getenv("AL_HOME")
	if alHome != "" {
		return alHome, nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(homeDir, ".al"), nil
}

// EnsureConfigDir creates the configuration directory if it doesn't exist
func EnsureConfigDir() error {
	configDir, err := GetConfigDir()
	if err != nil {
		return err
	}

	return os.MkdirAll(configDir, 0755)
}

// GetProvidersConfigPath returns the path to the providers.json file
func GetProvidersConfigPath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(configDir, "providers.json"), nil
}

// AppConfig represents the application configuration
type AppConfig struct {
	DefaultProvider string `json:"default_provider,omitempty"`
	DefaultProfile  string `json:"default_profile,omitempty"`
}

// GetConfigPath returns the path to the config.json file
func GetConfigPath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(configDir, "config.json"), nil
}

// LoadAppConfig loads the application configuration from JSON file
func LoadAppConfig() (*AppConfig, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	// If file doesn't exist, return empty config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &AppConfig{}, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config AppConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// SaveAppConfig saves the application configuration to JSON file
func SaveAppConfig(config *AppConfig) error {
	// Ensure config directory exists
	if err := EnsureConfigDir(); err != nil {
		return err
	}

	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

// SetDefaultProvider sets the default provider
func SetDefaultProvider(provider string) error {
	config, err := LoadAppConfig()
	if err != nil {
		return err
	}

	config.DefaultProvider = provider
	return SaveAppConfig(config)
}

// SetDefaultProfile sets the default profile
func SetDefaultProfile(profile string) error {
	config, err := LoadAppConfig()
	if err != nil {
		return err
	}

	config.DefaultProfile = profile
	return SaveAppConfig(config)
}
