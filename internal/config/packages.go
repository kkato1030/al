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
	ID          string    `json:"id"`           // required: brew="{formula,cask,tap}:<package_name>", mas="<app_id>"
	Name        string    `json:"name"`          // 表示用の名前（brewではidと同じ、masでは任意）
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

	// Check if package already exists with the same id, provider, and profile
	for _, existingPkg := range config.Packages {
		if existingPkg.ID == pkg.ID && existingPkg.Provider == pkg.Provider && existingPkg.Profile == pkg.Profile {
			return fmt.Errorf("package with id '%s' already exists for provider '%s' in profile '%s'", pkg.ID, pkg.Provider, pkg.Profile)
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

	// Check if package already exists with the same id, provider, and profile
	found := false
	for i, existingPkg := range config.Packages {
		if existingPkg.ID == pkg.ID && existingPkg.Provider == pkg.Provider && existingPkg.Profile == pkg.Profile {
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

// RemovePackage removes a package from the configuration
// Package is identified by id, provider, and profile combination
func RemovePackage(id, provider, profile string) error {
	config, err := LoadPackagesConfig()
	if err != nil {
		return err
	}

	// Find and remove the package
	found := false
	for i, pkg := range config.Packages {
		if pkg.ID == id && pkg.Provider == provider && pkg.Profile == profile {
			// Remove the package by creating a new slice without it
			config.Packages = append(config.Packages[:i], config.Packages[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("package with id '%s' with provider '%s' in profile '%s' not found", id, provider, profile)
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
