package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
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

// EnsureGitignore creates a .gitignore file in the config directory if it doesn't exist
func EnsureGitignore() error {
	configDir, err := GetConfigDir()
	if err != nil {
		return err
	}

	gitignorePath := filepath.Join(configDir, ".gitignore")
	
	// Check if .gitignore already exists
	if _, err := os.Stat(gitignorePath); err == nil {
		// File exists, check if logs/ is already in it
		content, err := os.ReadFile(gitignorePath)
		if err != nil {
			return err
		}
		
		// If logs/ is already present, don't modify
		contentStr := string(content)
		if strings.Contains(contentStr, "logs/") || strings.Contains(contentStr, "logs\n") {
			return nil
		}
		
		// Append logs/ to existing .gitignore
		f, err := os.OpenFile(gitignorePath, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		defer f.Close()
		
		// Add newline if file doesn't end with one
		if len(content) > 0 && content[len(content)-1] != '\n' {
			if _, err := f.WriteString("\n"); err != nil {
				return err
			}
		}
		
		if _, err := f.WriteString("logs/\n"); err != nil {
			return err
		}
		
		return nil
	}
	
	// Create new .gitignore
	content := "# Ignore execution logs\nlogs/\n"
	return os.WriteFile(gitignorePath, []byte(content), 0644)
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
	DefaultStage    string `json:"default_stage,omitempty"`
	// BackupRepo is the GitHub owner/repo for "al backup" (e.g. "kkato1030/dotal"). Empty means use default owner/dotal.
	BackupRepo string `json:"backup_repo,omitempty"`
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

// SetDefaultStage sets the default stage
func SetDefaultStage(stage string) error {
	config, err := LoadAppConfig()
	if err != nil {
		return err
	}

	config.DefaultStage = stage
	return SaveAppConfig(config)
}

// SetBackupRepo sets the backup repository (owner/repo) for "al backup"
func SetBackupRepo(repo string) error {
	config, err := LoadAppConfig()
	if err != nil {
		return err
	}

	config.BackupRepo = repo
	return SaveAppConfig(config)
}

// GetDefaultAliases returns the default command aliases
func GetDefaultAliases() map[string]string {
	return map[string]string{
		"promote":  "package move {args} --to package.promote_to",
		"add":      "package add",
		"remove":   "package remove",
		"list":     "package list",
		"register": "package add --provider manual",
		"import":   "package import {args}",
		"pkg":      "package",
		"prf":      "profile",
		"prv":      "provider",
	}
}
