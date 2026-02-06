package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// BrewTapsConfig holds the list of brew taps managed by al (provider brew).
// Taps are separate from packages; see https://github.com/kkato1030/al/issues/50
type BrewTapsConfig struct {
	Taps []string `json:"taps"`
}

// GetBrewTapsConfigPath returns the path to brew-taps.json under the config dir.
func GetBrewTapsConfigPath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "brew-taps.json"), nil
}

// LoadBrewTapsConfig loads brew taps from ~/.al/brew-taps.json.
func LoadBrewTapsConfig() (*BrewTapsConfig, error) {
	configPath, err := GetBrewTapsConfigPath()
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &BrewTapsConfig{Taps: []string{}}, nil
	}
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	var cfg BrewTapsConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if cfg.Taps == nil {
		cfg.Taps = []string{}
	}
	return &cfg, nil
}

// SaveBrewTapsConfig writes brew taps to ~/.al/brew-taps.json.
func SaveBrewTapsConfig(cfg *BrewTapsConfig) error {
	if err := EnsureConfigDir(); err != nil {
		return err
	}
	configPath, err := GetBrewTapsConfigPath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, data, 0644)
}

// AddBrewTap adds a tap name to the config if not already present.
func AddBrewTap(tapName string) error {
	tapName = strings.TrimSpace(tapName)
	if tapName == "" {
		return nil
	}
	cfg, err := LoadBrewTapsConfig()
	if err != nil {
		return err
	}
	for _, t := range cfg.Taps {
		if t == tapName {
			return nil
		}
	}
	cfg.Taps = append(cfg.Taps, tapName)
	return SaveBrewTapsConfig(cfg)
}

// RemoveBrewTap removes a tap name from the config.
func RemoveBrewTap(tapName string) error {
	cfg, err := LoadBrewTapsConfig()
	if err != nil {
		return err
	}
	for i, t := range cfg.Taps {
		if t == tapName {
			cfg.Taps = append(cfg.Taps[:i], cfg.Taps[i+1:]...)
			return SaveBrewTapsConfig(cfg)
		}
	}
	return nil
}

// HasBrewTap returns true if the tap is in the config.
func HasBrewTap(tapName string) (bool, error) {
	cfg, err := LoadBrewTapsConfig()
	if err != nil {
		return false, err
	}
	for _, t := range cfg.Taps {
		if t == tapName {
			return true, nil
		}
	}
	return false, nil
}
