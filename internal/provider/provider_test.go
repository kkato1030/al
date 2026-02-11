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
