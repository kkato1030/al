package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// PackageConfig represents a package configuration
type PackageConfig struct {
	Name        string    `json:"name"`
	Provider    string    `json:"provider"`
	Profile     string    `json:"profile"`
	Version     string    `json:"version,omitempty"`
	InstalledAt time.Time `json:"installed_at"`
	Description string    `json:"description,omitempty"`
}

// PackagesConfig represents the collection of package configurations
type PackagesConfig struct {
	Packages []PackageConfig `json:"packages"`
}

// LoadPackagesConfig loads the packages configuration from JSON file
func LoadPackagesConfig() (*PackagesConfig, error) {
	configPath, err := GetPackagesConfigPath()
	if err != nil {
		return nil, err
	}

	// If file doesn't exist, return empty config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &PackagesConfig{Packages: []PackageConfig{}}, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config PackagesConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// SavePackagesConfig saves the packages configuration to JSON file
func SavePackagesConfig(config *PackagesConfig) error {
	// Ensure config directory exists
	if err := EnsureConfigDir(); err != nil {
		return err
	}

	configPath, err := GetPackagesConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

// AddPackage adds a package to the configuration
// Returns an error if a package with the same name, provider, and profile already exists
func AddPackage(pkg PackageConfig) error {
	config, err := LoadPackagesConfig()
	if err != nil {
		return err
	}

	// Check if package already exists with the same name, provider, and profile
	for _, existingPkg := range config.Packages {
		if existingPkg.Name == pkg.Name && existingPkg.Provider == pkg.Provider && existingPkg.Profile == pkg.Profile {
			return fmt.Errorf("package '%s' already exists for provider '%s' in profile '%s'", pkg.Name, pkg.Provider, pkg.Profile)
		}
	}

	// Set InstalledAt if not set
	if pkg.InstalledAt.IsZero() {
		pkg.InstalledAt = time.Now()
	}

	// Add the package
	config.Packages = append(config.Packages, pkg)

	return SavePackagesConfig(config)
}

// AddOrUpdatePackage adds or updates a package in the configuration
func AddOrUpdatePackage(pkg PackageConfig) error {
	config, err := LoadPackagesConfig()
	if err != nil {
		return err
	}

	// Check if package already exists with the same name, provider, and profile
	found := false
	for i, existingPkg := range config.Packages {
		if existingPkg.Name == pkg.Name && existingPkg.Provider == pkg.Provider && existingPkg.Profile == pkg.Profile {
			// Update existing package
			// Preserve InstalledAt if not provided in new config
			if pkg.InstalledAt.IsZero() {
				pkg.InstalledAt = existingPkg.InstalledAt
			}
			config.Packages[i] = pkg
			found = true
			break
		}
	}

	// If not found, add it
	if !found {
		// Set InstalledAt if not set
		if pkg.InstalledAt.IsZero() {
			pkg.InstalledAt = time.Now()
		}
		config.Packages = append(config.Packages, pkg)
	}

	return SavePackagesConfig(config)
}

// GetPackagesConfigPath returns the path to the packages.json file
func GetPackagesConfigPath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(configDir, "packages.json"), nil
}
