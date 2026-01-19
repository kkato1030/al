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

// parsePackageID parses package ID in format "{formula,cask,tap}:<package_name>"
// Returns package type and name
func (p *BrewProvider) parsePackageID(packageID string) (pkgType, pkgName string, err error) {
	parts := strings.SplitN(packageID, ":", 2)
	if len(parts) != 2 {
		// If no prefix, default to formula
		return "formula", packageID, nil
	}

	pkgType = parts[0]
	pkgName = parts[1]

	if pkgType != "formula" && pkgType != "cask" && pkgType != "tap" {
		return "", "", fmt.Errorf("invalid package type: %s (must be formula, cask, or tap)", pkgType)
	}

	return pkgType, pkgName, nil
}

// detectPackageType detects if a package is a formula, cask, or tap
func (p *BrewProvider) detectPackageType(packageName string) (string, error) {
	// Try cask first (casks can have same name as formulas)
	cmd := exec.Command("brew", "info", "--cask", packageName)
	err := cmd.Run()
	if err == nil {
		return "cask", nil
	}

	// Try formula
	cmd = exec.Command("brew", "info", packageName)
	err = cmd.Run()
	if err == nil {
		return "formula", nil
	}

	// Try tap (format: tap_name/package_name or just tap_name)
	if strings.Contains(packageName, "/") {
		parts := strings.SplitN(packageName, "/", 2)
		if len(parts) == 2 {
			// Check if tap exists
			cmd = exec.Command("brew", "tap-info", parts[0])
			err = cmd.Run()
			if err == nil {
				return "tap", nil
			}
		}
	} else {
		// Check if it's a tap name itself
		cmd = exec.Command("brew", "tap-info", packageName)
		err = cmd.Run()
		if err == nil {
			return "tap", nil
		}
	}

	// Default to formula if cannot determine
	return "formula", nil
}

// GeneratePackageID generates package ID in format "{formula,cask,tap}:<package_name>"
func (p *BrewProvider) GeneratePackageID(packageName string) (string, error) {
	pkgType, err := p.detectPackageType(packageName)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s:%s", pkgType, packageName), nil
}

// InstallPackage installs a package using brew
// packageID is in format "{formula,cask,tap}:<package_name>"
func (p *BrewProvider) InstallPackage(packageID string) error {
	// Check if brew is installed
	installed, err := p.CheckInstalled()
	if err != nil {
		return fmt.Errorf("failed to check brew installation: %w", err)
	}
	if !installed {
		return fmt.Errorf("brew is not installed. Please install it first using 'al provider add brew'")
	}

	// Parse package ID
	pkgType, pkgName, err := p.parsePackageID(packageID)
	if err != nil {
		return fmt.Errorf("failed to parse package ID: %w", err)
	}

	// Run brew install command
	fmt.Printf("Installing %s using brew...\n", pkgName)
	var cmd *exec.Cmd
	if pkgType == "cask" {
		cmd = exec.Command("brew", "install", "--cask", pkgName)
	} else if pkgType == "tap" {
		cmd = exec.Command("brew", "tap", pkgName)
	} else {
		cmd = exec.Command("brew", "install", pkgName)
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install package %s: %w", pkgName, err)
	}

	fmt.Printf("Successfully installed %s\n", pkgName)
	return nil
}

// UninstallPackage uninstalls a package using brew
// packageID is in format "{formula,cask,tap}:<package_name>"
func (p *BrewProvider) UninstallPackage(packageID string) error {
	// Check if brew is installed
	installed, err := p.CheckInstalled()
	if err != nil {
		return fmt.Errorf("failed to check brew installation: %w", err)
	}
	if !installed {
		return fmt.Errorf("brew is not installed. Please install it first using 'al provider add brew'")
	}

	// Parse package ID
	pkgType, pkgName, err := p.parsePackageID(packageID)
	if err != nil {
		return fmt.Errorf("failed to parse package ID: %w", err)
	}

	// Run brew uninstall command
	fmt.Printf("Uninstalling %s using brew...\n", pkgName)
	var cmd *exec.Cmd
	if pkgType == "cask" {
		cmd = exec.Command("brew", "uninstall", "--cask", pkgName)
	} else if pkgType == "tap" {
		cmd = exec.Command("brew", "untap", pkgName)
	} else {
		cmd = exec.Command("brew", "uninstall", pkgName)
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to uninstall package %s: %w", pkgName, err)
	}

	fmt.Printf("Successfully uninstalled %s\n", pkgName)
	return nil
}

// UpgradePackage upgrades a package using brew
// packageID is in format "{formula,cask,tap}:<package_name>"
func (p *BrewProvider) UpgradePackage(packageID string) error {
	// Check if brew is installed
	installed, err := p.CheckInstalled()
	if err != nil {
		return fmt.Errorf("failed to check brew installation: %w", err)
	}
	if !installed {
		return fmt.Errorf("brew is not installed. Please install it first using 'al provider add brew'")
	}

	// Parse package ID
	pkgType, pkgName, err := p.parsePackageID(packageID)
	if err != nil {
		return fmt.Errorf("failed to parse package ID: %w", err)
	}

	// Run brew upgrade command
	fmt.Printf("Upgrading %s using brew...\n", pkgName)
	var cmd *exec.Cmd
	if pkgType == "cask" {
		cmd = exec.Command("brew", "upgrade", "--cask", pkgName)
	} else if pkgType == "tap" {
		// Taps don't have upgrade, but we can reinstall
		cmd = exec.Command("brew", "tap", pkgName)
	} else {
		cmd = exec.Command("brew", "upgrade", pkgName)
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to upgrade package %s: %w", pkgName, err)
	}

	fmt.Printf("Successfully upgraded %s\n", pkgName)
	return nil
}

// Upgrade upgrades brew itself
func (p *BrewProvider) Upgrade() error {
	// Check if brew is installed
	installed, err := p.CheckInstalled()
	if err != nil {
		return fmt.Errorf("failed to check brew installation: %w", err)
	}
	if !installed {
		return fmt.Errorf("brew is not installed. Please install it first using 'al provider add brew'")
	}

	// Run brew update and upgrade
	fmt.Println("Updating brew...")
	updateCmd := exec.Command("brew", "update")
	updateCmd.Stdin = os.Stdin
	updateCmd.Stdout = os.Stdout
	updateCmd.Stderr = os.Stderr

	if err := updateCmd.Run(); err != nil {
		return fmt.Errorf("failed to update brew: %w", err)
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

	fmt.Println("Successfully upgraded brew")
	return nil
}

// SearchPackage searches for packages using brew search
func (p *BrewProvider) SearchPackage(query string) ([]SearchResult, error) {
	// Check if brew is installed
	installed, err := p.CheckInstalled()
	if err != nil {
		return nil, fmt.Errorf("failed to check brew installation: %w", err)
	}
	if !installed {
		return nil, fmt.Errorf("brew is not installed. Please install it first using 'al provider add brew'")
	}

	// Run brew search command
	cmd := exec.Command("brew", "search", query)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to search packages: %w", err)
	}

	// Parse output - brew search returns one package name per line
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	results := make([]SearchResult, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Detect package type and generate ID
		pkgType, err := p.detectPackageType(line)
		if err != nil {
			// If detection fails, default to formula
			pkgType = "formula"
		}

		pkgID := fmt.Sprintf("%s:%s", pkgType, line)
		results = append(results, SearchResult{
			ID:          pkgID,
			Name:        line,
			Description: "",
		})
	}

	return results, nil
}
