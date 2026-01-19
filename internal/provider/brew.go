package provider

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/kkato1030/al/internal/config"
)

// BrewProvider implements the Provider interface for Homebrew
type BrewProvider struct {
	name string
}

// NewBrewProvider creates a new brew provider
func NewBrewProvider() *BrewProvider {
	return &BrewProvider{name: "brew"}
}

// Name returns the provider name
func (p *BrewProvider) Name() string {
	return p.name
}

// CheckInstalled checks if brew is installed by running `brew --version`
func (p *BrewProvider) CheckInstalled() (bool, error) {
	cmd := exec.Command("brew", "--version")
	err := cmd.Run()
	if err != nil {
		// If command fails, brew is not installed
		return false, nil
	}
	return true, nil
}

// GetVersion returns the version of brew
func (p *BrewProvider) GetVersion() (string, error) {
	cmd := exec.Command("brew", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	// Parse version from output (e.g., "Homebrew 4.x.x")
	lines := strings.Split(string(output), "\n")
	if len(lines) > 0 {
		parts := strings.Fields(lines[0])
		if len(parts) >= 2 {
			return parts[1], nil
		}
	}

	return strings.TrimSpace(string(output)), nil
}

// Install installs Homebrew using the official installation script
func (p *BrewProvider) Install() error {
	// Check if brew is already installed
	installed, err := p.CheckInstalled()
	if err != nil {
		return fmt.Errorf("failed to check brew installation: %w", err)
	}
	if installed {
		return fmt.Errorf("brew is already installed")
	}

	// Run the official Homebrew installation script
	installScript := "/bin/bash -c \"$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)\""
	cmd := exec.Command("sh", "-c", installScript)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install brew: %w", err)
	}

	return nil
}

// SetupConfig sets up the configuration for brew provider
func (p *BrewProvider) SetupConfig() error {
	// Ensure config directory exists
	if err := config.EnsureConfigDir(); err != nil {
		return fmt.Errorf("failed to ensure config directory: %w", err)
	}

	// Get version
	version, err := p.GetVersion()
	if err != nil {
		// If version cannot be retrieved, continue without it
		version = ""
	}

	// Add provider to config
	providerConfig := config.ProviderConfig{
		Name:        p.name,
		InstalledAt: time.Now(),
		Version:     version,
	}

	if err := config.AddOrUpdateProvider(providerConfig); err != nil {
		return fmt.Errorf("failed to save provider config: %w", err)
	}

	return nil
}

// InstallPackage installs a package using brew
func (p *BrewProvider) InstallPackage(packageName string) error {
	// Check if brew is installed
	installed, err := p.CheckInstalled()
	if err != nil {
		return fmt.Errorf("failed to check brew installation: %w", err)
	}
	if !installed {
		return fmt.Errorf("brew is not installed. Please install it first using 'al provider add brew'")
	}

	// Run brew install command
	fmt.Printf("Installing %s using brew...\n", packageName)
	cmd := exec.Command("brew", "install", packageName)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install package %s: %w", packageName, err)
	}

	fmt.Printf("Successfully installed %s\n", packageName)
	return nil
}
