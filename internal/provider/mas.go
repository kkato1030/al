package provider

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/kkato1030/al/internal/config"
)

// MasProvider implements the Provider interface for Mac App Store (mas)
type MasProvider struct {
	name string
}

// NewMasProvider creates a new mas provider
func NewMasProvider() *MasProvider {
	return &MasProvider{name: "mas"}
}

// Name returns the provider name
func (p *MasProvider) Name() string {
	return p.name
}

// CheckInstalled checks if mas is installed by running `mas version`
func (p *MasProvider) CheckInstalled() (bool, error) {
	cmd := exec.Command("mas", "version")
	err := cmd.Run()
	if err != nil {
		// If command fails, mas is not installed
		return false, nil
	}
	return true, nil
}

// GetVersion returns the version of mas
func (p *MasProvider) GetVersion() (string, error) {
	cmd := exec.Command("mas", "version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

// Install installs mas using Homebrew
func (p *MasProvider) Install() error {
	// Check if mas is already installed
	installed, err := p.CheckInstalled()
	if err != nil {
		return fmt.Errorf("failed to check mas installation: %w", err)
	}
	if installed {
		return fmt.Errorf("mas is already installed")
	}

	// Check if brew is installed first
	brewInstalled := false
	brewCmd := exec.Command("brew", "--version")
	if err := brewCmd.Run(); err == nil {
		brewInstalled = true
	}

	if !brewInstalled {
		return fmt.Errorf("brew is required to install mas. Please install brew first using 'al provider add brew'")
	}

	// Install mas using brew
	fmt.Println("Installing mas using brew...")
	cmd := exec.Command("brew", "install", "mas")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install mas: %w", err)
	}

	return nil
}

// SetupConfig sets up the configuration for mas provider
func (p *MasProvider) SetupConfig() error {
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

// InstallPackage installs a package using mas
// packageID is the app ID (id = app_id for mas)
func (p *MasProvider) InstallPackage(packageID string) error {
	// Check if mas is installed
	installed, err := p.CheckInstalled()
	if err != nil {
		return fmt.Errorf("failed to check mas installation: %w", err)
	}
	if !installed {
		return fmt.Errorf("mas is not installed. Please install it first using 'al provider add mas'")
	}

	// Run mas install command
	fmt.Printf("Installing %s using mas...\n", packageID)
	cmd := exec.Command("mas", "install", packageID)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install package %s: %w", packageID, err)
	}

	fmt.Printf("Successfully installed %s\n", packageID)
	return nil
}

// UninstallPackage uninstalls a package using mas
// packageID is the app ID (id = app_id for mas)
func (p *MasProvider) UninstallPackage(packageID string) error {
	// Check if mas is installed
	installed, err := p.CheckInstalled()
	if err != nil {
		return fmt.Errorf("failed to check mas installation: %w", err)
	}
	if !installed {
		return fmt.Errorf("mas is not installed. Please install it first using 'al provider add mas'")
	}

	// Run mas uninstall command
	fmt.Printf("Uninstalling %s using mas...\n", packageID)
	cmd := exec.Command("mas", "uninstall", packageID)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to uninstall package %s: %w", packageID, err)
	}

	fmt.Printf("Successfully uninstalled %s\n", packageID)
	return nil
}

// UpgradePackage upgrades a package using mas
// packageID is the app ID (id = app_id for mas)
func (p *MasProvider) UpgradePackage(packageID string) error {
	// Check if mas is installed
	installed, err := p.CheckInstalled()
	if err != nil {
		return fmt.Errorf("failed to check mas installation: %w", err)
	}
	if !installed {
		return fmt.Errorf("mas is not installed. Please install it first using 'al provider add mas'")
	}

	// Run mas upgrade command
	fmt.Printf("Upgrading %s using mas...\n", packageID)
	cmd := exec.Command("mas", "upgrade", packageID)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to upgrade package %s: %w", packageID, err)
	}

	fmt.Printf("Successfully upgraded %s\n", packageID)
	return nil
}

// Upgrade upgrades mas itself using brew
func (p *MasProvider) Upgrade() error {
	// Check if mas is installed
	installed, err := p.CheckInstalled()
	if err != nil {
		return fmt.Errorf("failed to check mas installation: %w", err)
	}
	if !installed {
		return fmt.Errorf("mas is not installed. Please install it first using 'al provider add mas'")
	}

	// Check if brew is installed
	brewInstalled := false
	brewCmd := exec.Command("brew", "--version")
	if err := brewCmd.Run(); err == nil {
		brewInstalled = true
	}

	if !brewInstalled {
		return fmt.Errorf("brew is required to upgrade mas. Please install brew first using 'al provider add brew'")
	}

	// Run brew upgrade mas
	fmt.Println("Upgrading mas using brew...")
	cmd := exec.Command("brew", "upgrade", "mas")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to upgrade mas: %w", err)
	}

	// Update version in config
	version, err := p.GetVersion()
	if err == nil {
		providerConfig := config.ProviderConfig{
			Name:        p.name,
			InstalledAt: time.Now(),
			Version:     version,
		}
		if err := config.AddOrUpdateProvider(providerConfig); err != nil {
			fmt.Printf("Warning: failed to update provider config: %v\n", err)
		}
	}

	fmt.Println("Successfully upgraded mas")
	return nil
}

// SearchPackage searches for packages using mas search
func (p *MasProvider) SearchPackage(query string) ([]SearchResult, error) {
	// Check if mas is installed
	installed, err := p.CheckInstalled()
	if err != nil {
		return nil, fmt.Errorf("failed to check mas installation: %w", err)
	}
	if !installed {
		return nil, fmt.Errorf("mas is not installed. Please install it first using 'al provider add mas'")
	}

	// Run mas search command
	cmd := exec.Command("mas", "search", query)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to search packages: %w", err)
	}

	// Parse output - mas search returns lines like:
	// "123456789 App Name (Category)"
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	results := make([]SearchResult, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse line: "123456789 App Name (Category)"
		// Extract app ID (first number) and name
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		appID := parts[0]
		// Name is everything after the ID, but remove parentheses if present
		nameParts := parts[1:]
		name := strings.Join(nameParts, " ")
		// Remove trailing parentheses if present
		if strings.HasSuffix(name, ")") {
			lastParen := strings.LastIndex(name, "(")
			if lastParen > 0 {
				name = strings.TrimSpace(name[:lastParen])
			}
		}

		results = append(results, SearchResult{
			ID:          appID,
			Name:        name,
			Description: "",
		})
	}

	return results, nil
}
