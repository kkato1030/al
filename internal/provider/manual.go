package provider

import (
	"fmt"
	"time"

	"github.com/kkato1030/al/internal/config"
)

// ManualProvider implements the Provider interface for manually installed packages
// This provider is used to track packages that are installed manually (via shell scripts, pkg/dmg files, etc.)
type ManualProvider struct {
	name string
}

// NewManualProvider creates a new manual provider
func NewManualProvider() *ManualProvider {
	return &ManualProvider{name: "manual"}
}

// Name returns the provider name
func (p *ManualProvider) Name() string {
	return p.name
}

// CheckInstalled always returns true for manual provider
// Manual provider is always available as it's just a tracking mechanism
func (p *ManualProvider) CheckInstalled() (bool, error) {
	return true, nil
}

// Install does nothing for manual provider
// Manual provider doesn't need installation
func (p *ManualProvider) Install() error {
	// Manual provider doesn't require installation
	return nil
}

// SetupConfig sets up the configuration for manual provider
func (p *ManualProvider) SetupConfig() error {
	// Ensure config directory exists
	if err := config.EnsureConfigDir(); err != nil {
		return fmt.Errorf("failed to ensure config directory: %w", err)
	}

	// Add provider to config
	providerConfig := config.ProviderConfig{
		Name:        p.name,
		InstalledAt: time.Now(),
		Version:     "", // Manual provider doesn't have a version
	}

	if err := config.AddOrUpdateProvider(providerConfig); err != nil {
		return fmt.Errorf("failed to save provider config: %w", err)
	}

	return nil
}

// InstallPackage does nothing for manual provider
// Packages are assumed to be already installed manually
func (p *ManualProvider) InstallPackage(packageID string) error {
	// Manual provider doesn't install packages
	// Packages are assumed to be already installed manually
	fmt.Printf("Note: Package '%s' is tracked as manually installed. Please ensure it is already installed.\n", packageID)
	return nil
}

// UninstallPackage does nothing for manual provider
// Users need to uninstall packages manually
func (p *ManualProvider) UninstallPackage(packageID string) error {
	// Manual provider doesn't uninstall packages
	// Users need to uninstall packages manually
	fmt.Printf("Note: Package '%s' is tracked as manually installed. Please uninstall it manually if needed.\n", packageID)
	return nil
}

// UpgradePackage does nothing for manual provider
// Users need to upgrade packages manually
func (p *ManualProvider) UpgradePackage(packageID string) error {
	// Manual provider doesn't upgrade packages
	// Users need to upgrade packages manually
	fmt.Printf("Note: Package '%s' is tracked as manually installed. Please upgrade it manually if needed.\n", packageID)
	return nil
}

// Upgrade does nothing for manual provider
func (p *ManualProvider) Upgrade() error {
	// Manual provider doesn't have an upgrade mechanism
	return nil
}

// SearchPackage returns an empty result for manual provider
// Manual provider doesn't support searching
func (p *ManualProvider) SearchPackage(query string) ([]SearchResult, error) {
	// Manual provider doesn't support searching
	return []SearchResult{}, nil
}
