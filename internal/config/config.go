package config

import (
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
