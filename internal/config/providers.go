package config

import (
	"encoding/json"
	"os"
	"time"
)

// ProviderConfig represents a provider configuration
type ProviderConfig struct {
	Name        string    `json:"name"`
	InstalledAt time.Time `json:"installed_at"`
	Version     string    `json:"version,omitempty"`
}

// ProvidersConfig represents the collection of provider configurations
type ProvidersConfig struct {
	Providers []ProviderConfig `json:"providers"`
}

// LoadProvidersConfig loads the providers configuration from JSON file
func LoadProvidersConfig() (*ProvidersConfig, error) {
	configPath, err := GetProvidersConfigPath()
	if err != nil {
		return nil, err
	}

	// If file doesn't exist, return empty config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &ProvidersConfig{Providers: []ProviderConfig{}}, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config ProvidersConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// SaveProvidersConfig saves the providers configuration to JSON file
func SaveProvidersConfig(config *ProvidersConfig) error {
	// Ensure config directory exists
	if err := EnsureConfigDir(); err != nil {
		return err
	}

	configPath, err := GetProvidersConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

// AddOrUpdateProvider adds or updates a provider in the configuration
func AddOrUpdateProvider(provider ProviderConfig) error {
	config, err := LoadProvidersConfig()
	if err != nil {
		return err
	}

	// Check if provider already exists
	found := false
	for i, p := range config.Providers {
		if p.Name == provider.Name {
			config.Providers[i] = provider
			found = true
			break
		}
	}

	// If not found, add it
	if !found {
		config.Providers = append(config.Providers, provider)
	}

	return SaveProvidersConfig(config)
}

// GetProvider returns a provider configuration by name
func GetProvider(name string) (*ProviderConfig, error) {
	config, err := LoadProvidersConfig()
	if err != nil {
		return nil, err
	}

	for _, p := range config.Providers {
		if p.Name == name {
			return &p, nil
		}
	}

	return nil, nil // Provider not found
}
