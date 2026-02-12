package provider

import (
	"testing"
)

// TestBrewProvider_IsPackageInstalled tests the IsPackageInstalled method
// This test will be skipped if brew is not installed
func TestBrewProvider_IsPackageInstalled(t *testing.T) {
	provider := NewBrewProvider()
	installed, err := provider.CheckInstalled()
	if err != nil {
		t.Fatalf("Failed to check if brew is installed: %v", err)
	}
	if !installed {
		t.Skip("Skipping test: brew is not installed")
	}

	// Test with a commonly installed formula (brew itself uses git)
	tests := []struct {
		name      string
		packageID string
	}{
		{
			name:      "formula package",
			packageID: "formula:git",
		},
		{
			name:      "package without prefix",
			packageID: "git",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			installed, err := provider.IsPackageInstalled(tt.packageID)
			if err != nil {
				t.Errorf("IsPackageInstalled(%q) error = %v", tt.packageID, err)
			}
			// We can't assert the result since we don't know what's installed
			// Just ensure no error is returned
			t.Logf("IsPackageInstalled(%q) = %v", tt.packageID, installed)
		})
	}
}

// TestBrewProvider_IsPackageUpgradable tests the IsPackageUpgradable method
// This test will be skipped if brew is not installed
func TestBrewProvider_IsPackageUpgradable(t *testing.T) {
	provider := NewBrewProvider()
	installed, err := provider.CheckInstalled()
	if err != nil {
		t.Fatalf("Failed to check if brew is installed: %v", err)
	}
	if !installed {
		t.Skip("Skipping test: brew is not installed")
	}

	// Test with a package ID (whether it's upgradable depends on current state)
	packageID := "formula:git"
	upgradable, err := provider.IsPackageUpgradable(packageID)
	if err != nil {
		t.Errorf("IsPackageUpgradable(%q) error = %v", packageID, err)
	}
	// We can't assert the result since it depends on system state
	// Just ensure no error is returned
	t.Logf("IsPackageUpgradable(%q) = %v", packageID, upgradable)
}

// TestMasProvider_IsPackageInstalled tests the IsPackageInstalled method
// This test will be skipped if mas is not installed
func TestMasProvider_IsPackageInstalled(t *testing.T) {
	provider := NewMasProvider()
	installed, err := provider.CheckInstalled()
	if err != nil {
		t.Fatalf("Failed to check if mas is installed: %v", err)
	}
	if !installed {
		t.Skip("Skipping test: mas is not installed")
	}

	// Test with a fake app ID
	packageID := "999999999"
	installed, err = provider.IsPackageInstalled(packageID)
	if err != nil {
		t.Errorf("IsPackageInstalled(%q) error = %v", packageID, err)
	}
	// Should not be installed
	if installed {
		t.Errorf("IsPackageInstalled(%q) = true, want false", packageID)
	}
}

// TestManualProvider_Methods tests the manual provider methods
func TestManualProvider_Methods(t *testing.T) {
	provider := NewManualProvider()

	// Manual provider should always be "installed"
	installed, err := provider.CheckInstalled()
	if err != nil {
		t.Errorf("CheckInstalled() error = %v", err)
	}
	if !installed {
		t.Errorf("CheckInstalled() = false, want true")
	}

	// IsPackageInstalled should return false for manual packages
	packageID := "test-package"
	pkgInstalled, err := provider.IsPackageInstalled(packageID)
	if err != nil {
		t.Errorf("IsPackageInstalled(%q) error = %v", packageID, err)
	}
	if pkgInstalled {
		t.Errorf("IsPackageInstalled(%q) = true, want false", packageID)
	}

	// IsPackageUpgradable should return false for manual packages
	upgradable, err := provider.IsPackageUpgradable(packageID)
	if err != nil {
		t.Errorf("IsPackageUpgradable(%q) error = %v", packageID, err)
	}
	if upgradable {
		t.Errorf("IsPackageUpgradable(%q) = true, want false", packageID)
	}
}

// TestBrewProvider_ListInstalled tests the ListInstalled batch method
func TestBrewProvider_ListInstalled(t *testing.T) {
	provider := NewBrewProvider()
	providerInstalled, err := provider.CheckInstalled()
	if err != nil {
		t.Fatalf("Failed to check if brew is installed: %v", err)
	}
	if !providerInstalled {
		t.Skip("Skipping test: brew is not installed")
	}

	packages, err := provider.ListInstalled()
	if err != nil {
		t.Errorf("ListInstalled() error = %v", err)
	}
	// Just ensure it returns without error
	t.Logf("ListInstalled() returned %d packages", len(packages))
}

// TestBrewProvider_ListUpgradable tests the ListUpgradable batch method
func TestBrewProvider_ListUpgradable(t *testing.T) {
	provider := NewBrewProvider()
	providerInstalled, err := provider.CheckInstalled()
	if err != nil {
		t.Fatalf("Failed to check if brew is installed: %v", err)
	}
	if !providerInstalled {
		t.Skip("Skipping test: brew is not installed")
	}

	packages, err := provider.ListUpgradable()
	if err != nil {
		t.Errorf("ListUpgradable() error = %v", err)
	}
	// Just ensure it returns without error
	t.Logf("ListUpgradable() returned %d packages", len(packages))
}

// TestMasProvider_ListInstalled tests the ListInstalled batch method
func TestMasProvider_ListInstalled(t *testing.T) {
	provider := NewMasProvider()
	providerInstalled, err := provider.CheckInstalled()
	if err != nil {
		t.Fatalf("Failed to check if mas is installed: %v", err)
	}
	if !providerInstalled {
		t.Skip("Skipping test: mas is not installed")
	}

	packages, err := provider.ListInstalled()
	if err != nil {
		t.Errorf("ListInstalled() error = %v", err)
	}
	// Just ensure it returns without error
	t.Logf("ListInstalled() returned %d packages", len(packages))
}

// TestMasProvider_ListUpgradable tests the ListUpgradable batch method
func TestMasProvider_ListUpgradable(t *testing.T) {
	provider := NewMasProvider()
	providerInstalled, err := provider.CheckInstalled()
	if err != nil {
		t.Fatalf("Failed to check if mas is installed: %v", err)
	}
	if !providerInstalled {
		t.Skip("Skipping test: mas is not installed")
	}

	packages, err := provider.ListUpgradable()
	if err != nil {
		t.Errorf("ListUpgradable() error = %v", err)
	}
	// Just ensure it returns without error
	t.Logf("ListUpgradable() returned %d packages", len(packages))
}

// TestManualProvider_BatchMethods tests the batch methods for manual provider
func TestManualProvider_BatchMethods(t *testing.T) {
	provider := NewManualProvider()

	// ListInstalled should return empty map
	installed, err := provider.ListInstalled()
	if err != nil {
		t.Errorf("ListInstalled() error = %v", err)
	}
	if len(installed) != 0 {
		t.Errorf("ListInstalled() returned %d packages, want 0", len(installed))
	}

	// ListUpgradable should return empty map
	upgradable, err := provider.ListUpgradable()
	if err != nil {
		t.Errorf("ListUpgradable() error = %v", err)
	}
	if len(upgradable) != 0 {
		t.Errorf("ListUpgradable() returned %d packages, want 0", len(upgradable))
	}
}

// TestBrewProvider_detectPackageType tests the package type detection logic
func TestBrewProvider_detectPackageType(t *testing.T) {
	provider := NewBrewProvider()

	tests := []struct {
		name        string
		packageName string
		wantType    string
	}{
		{
			name:        "formula from tap with 3 parts",
			packageName: "entireio/tap/entire",
			wantType:    "formula",
		},
		{
			name:        "tap with 2 parts",
			packageName: "entireio/tap",
			wantType:    "tap",
		},
		{
			name:        "simple formula",
			packageName: "git",
			wantType:    "formula",
		},
		{
			name:        "formula from tap with owner/repo/formula format",
			packageName: "homebrew/cask-versions/firefox-esr",
			wantType:    "formula",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotType, err := provider.detectPackageType(tt.packageName)
			if err != nil {
				t.Errorf("detectPackageType(%q) error = %v", tt.packageName, err)
				return
			}
			if gotType != tt.wantType {
				t.Errorf("detectPackageType(%q) = %q, want %q", tt.packageName, gotType, tt.wantType)
			}
		})
	}
}
