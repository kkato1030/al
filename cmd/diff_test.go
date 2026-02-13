package cmd

import (
	"testing"
)

func TestCalculateDiff_TapFormulas(t *testing.T) {
	// Test case: tap formulas should match their short names in installed packages
	desired := map[string]packageDiff{
		"formula:entireio/tap/entire": {
			provider: "brew",
			name:     "entire",
			id:       "formula:entireio/tap/entire",
		},
		"formula:kkato1030/tap/mm": {
			provider: "brew",
			name:     "mm",
			id:       "formula:kkato1030/tap/mm",
		},
		"formula:git": {
			provider: "brew",
			name:     "git",
			id:       "formula:git",
		},
	}

	// Simulate what brew list --installed-on-request returns
	// It returns short names for all formulas, including tap formulas
	installed := map[string]bool{
		"entire":         true, // short name from tap
		"formula:entire": true, // we store both
		"mm":             true, // short name from tap
		"formula:mm":     true, // we store both
		"git":            true, // regular formula
		"formula:git":    true, // we store both
		"curl":           true, // not in desired (removal)
		"formula:curl":   true,
	}

	upgradable := map[string]bool{}

	result := calculateDiff(desired, installed, upgradable)

	// Should have no additions (all desired packages are installed)
	if len(result.additions) != 0 {
		t.Errorf("Expected 0 additions, got %d: %+v", len(result.additions), result.additions)
	}

	// Should have 1 removal (curl)
	if len(result.removals) != 1 {
		t.Errorf("Expected 1 removal, got %d: %+v", len(result.removals), result.removals)
		for _, r := range result.removals {
			t.Logf("  Removal: %s (id: %s)", r.name, r.id)
		}
	} else if result.removals[0].name != "curl" {
		t.Errorf("Expected removal 'curl', got '%s'", result.removals[0].name)
	}

	// Should have no upgrades
	if len(result.upgrades) != 0 {
		t.Errorf("Expected 0 upgrades, got %d: %+v", len(result.upgrades), result.upgrades)
	}
}

func TestCalculateDiff_RegularFormulas(t *testing.T) {
	// Test case: regular formulas without tap
	desired := map[string]packageDiff{
		"formula:git": {
			provider: "brew",
			name:     "git",
			id:       "formula:git",
		},
		"formula:jq": {
			provider: "brew",
			name:     "jq",
			id:       "formula:jq",
		},
	}

	installed := map[string]bool{
		"git":         true,
		"formula:git": true,
		"jq":          true,
		"formula:jq":  true,
	}

	upgradable := map[string]bool{}

	result := calculateDiff(desired, installed, upgradable)

	// Should have no additions
	if len(result.additions) != 0 {
		t.Errorf("Expected 0 additions, got %d: %+v", len(result.additions), result.additions)
	}

	// Should have no removals (all installed packages are in desired)
	if len(result.removals) != 0 {
		t.Errorf("Expected 0 removals, got %d: %+v", len(result.removals), result.removals)
	}

	// Should have no upgrades
	if len(result.upgrades) != 0 {
		t.Errorf("Expected 0 upgrades, got %d: %+v", len(result.upgrades), result.upgrades)
	}
}

func TestCalculateDiff_UpgradableTapFormula(t *testing.T) {
	// Test case: tap formula that has an upgrade available
	desired := map[string]packageDiff{
		"formula:entireio/tap/entire": {
			provider: "brew",
			name:     "entire",
			id:       "formula:entireio/tap/entire",
		},
	}

	installed := map[string]bool{
		"entire":         true,
		"formula:entire": true,
	}

	// brew outdated returns short names
	upgradable := map[string]bool{
		"entire":         true,
		"formula:entire": true,
	}

	result := calculateDiff(desired, installed, upgradable)

	// Should have no additions
	if len(result.additions) != 0 {
		t.Errorf("Expected 0 additions, got %d: %+v", len(result.additions), result.additions)
	}

	// Should have no removals
	if len(result.removals) != 0 {
		t.Errorf("Expected 0 removals, got %d: %+v", len(result.removals), result.removals)
	}

	// Should have 1 upgrade for the tap formula
	if len(result.upgrades) != 1 {
		t.Errorf("Expected 1 upgrade, got %d: %+v", len(result.upgrades), result.upgrades)
	} else if result.upgrades[0].id != "formula:entireio/tap/entire" {
		t.Errorf("Expected upgrade 'formula:entireio/tap/entire', got '%s'", result.upgrades[0].id)
	}
}

func TestCalculateDiff_TapCasks(t *testing.T) {
	// Test case: tap casks should match their short names in installed packages
	// This handles cases like entireio/tap/entire which is a cask from a tap
	desired := map[string]packageDiff{
		"formula:entireio/tap/entire": {
			provider: "brew",
			name:     "entire",
			id:       "formula:entireio/tap/entire",
		},
		"cask:somevendor/tap/sometool": {
			provider: "brew",
			name:     "sometool",
			id:       "cask:somevendor/tap/sometool",
		},
	}

	// Simulate what brew list returns
	// The tap package might be installed as a cask even if desired as formula
	installed := map[string]bool{
		"entire":        true, // short name from tap (installed as cask)
		"cask:entire":   true, // we store both
		"sometool":      true, // short name from tap cask
		"cask:sometool": true, // we store both
	}

	upgradable := map[string]bool{}

	result := calculateDiff(desired, installed, upgradable)

	// Should have no additions (packages are installed, even if type differs)
	if len(result.additions) != 0 {
		t.Errorf("Expected 0 additions, got %d: %+v", len(result.additions), result.additions)
	}

	// Should have no removals (all installed packages map to desired tap packages)
	if len(result.removals) != 0 {
		t.Errorf("Expected 0 removals, got %d: %+v", len(result.removals), result.removals)
	}

	// Should have no upgrades
	if len(result.upgrades) != 0 {
		t.Errorf("Expected 0 upgrades, got %d: %+v", len(result.upgrades), result.upgrades)
	}
}
