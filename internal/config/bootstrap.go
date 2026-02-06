package config

import (
	"fmt"
	"os"
	"path/filepath"
)

const bootstrapScriptName = "script.sh"

// DefaultBootstrapScriptContent is the initial content when creating the bootstrap script.
const DefaultBootstrapScriptContent = "#!/usr/bin/env bash\n# Bootstrap script for new PC setup. Edit with: al bootstrap edit\nset -euo pipefail\n\n"

// GetBootstrapDir returns the path to ~/.al/bootstrap/
func GetBootstrapDir() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "bootstrap"), nil
}

// GetBootstrapScriptPath returns the path to ~/.al/bootstrap/script.sh
func GetBootstrapScriptPath() (string, error) {
	dir, err := GetBootstrapDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, bootstrapScriptName), nil
}

// BootstrapScriptExists returns true if the bootstrap script file exists.
func BootstrapScriptExists() (bool, error) {
	path, err := GetBootstrapScriptPath()
	if err != nil {
		return false, err
	}
	_, err = os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// EnsureBootstrapScript creates the bootstrap directory and script.sh with default content if it does not exist.
// Returns the script path and whether the file was created (true) or already existed (false).
func EnsureBootstrapScript() (path string, created bool, err error) {
	dir, err := GetBootstrapDir()
	if err != nil {
		return "", false, err
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", false, fmt.Errorf("creating bootstrap dir: %w", err)
	}
	path = filepath.Join(dir, bootstrapScriptName)
	_, statErr := os.Stat(path)
	if statErr == nil {
		return path, false, nil
	}
	if !os.IsNotExist(statErr) {
		return "", false, statErr
	}
	if err := os.WriteFile(path, []byte(DefaultBootstrapScriptContent), 0755); err != nil {
		return "", false, fmt.Errorf("writing bootstrap script: %w", err)
	}
	return path, true, nil
}

// ReadBootstrapScript returns the content of the bootstrap script, or error if it does not exist.
func ReadBootstrapScript() ([]byte, error) {
	path, err := GetBootstrapScriptPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("bootstrap script not found (run 'al bootstrap add' to create)")
		}
		return nil, err
	}
	return data, nil
}

// RemoveBootstrapScript deletes the bootstrap script file.
// The bootstrap directory is left in place (empty or not).
func RemoveBootstrapScript() error {
	path, err := GetBootstrapScriptPath()
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("bootstrap script not found")
		}
		return err
	}
	return nil
}
