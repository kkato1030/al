package config

import (
	"testing"
	"time"
)

// TestGetOverduePackages_WithReviewByPastDate tests the scenario from issue #59:
// A package with review_by set to a past date should be detected as overdue
// if the profile has review_days set.
func TestGetOverduePackages_WithReviewByPastDate(t *testing.T) {
	dir := t.TempDir()
	restore := setEnv("AL_HOME", dir)
	defer restore()

	if err := EnsureConfigDir(); err != nil {
		t.Fatal(err)
	}

	// Create profile with review_days (like the trial profile in stable-trial template)
	days30 := 30
	if err := SaveProfilesConfig(&ProfilesConfig{
		Profiles: []ProfileConfig{
			{Name: "test.trial", Stage: "trial", ReviewDays: &days30},
		},
	}); err != nil {
		t.Fatal(err)
	}

	// Add package with review_by set to past date (Feb 8, simulating issue scenario)
	pastDate := time.Date(2026, 2, 8, 0, 0, 0, 0, time.UTC)
	pkg := PackageConfig{
		ID:          "formula:test-pkg",
		Name:        "test-pkg",
		Provider:    "brew",
		Profile:     "test.trial",
		InstalledAt: time.Now(),
		ReviewBy:    &pastDate,
	}
	if err := AddPackage(pkg); err != nil {
		t.Fatal(err)
	}

	// GetOverduePackages should return this package
	overdue, err := GetOverduePackages()
	if err != nil {
		t.Fatal(err)
	}

	if len(overdue) != 1 {
		t.Errorf("Expected 1 overdue package, got %d", len(overdue))
	}
	if len(overdue) > 0 && overdue[0].ID != "formula:test-pkg" {
		t.Errorf("Expected overdue package to be 'formula:test-pkg', got %s", overdue[0].ID)
	}
}

// TestGetOverduePackages_WithoutReviewBy tests that packages without review_by
// are considered overdue if the profile has review_days.
func TestGetOverduePackages_WithoutReviewBy(t *testing.T) {
	dir := t.TempDir()
	restore := setEnv("AL_HOME", dir)
	defer restore()

	if err := EnsureConfigDir(); err != nil {
		t.Fatal(err)
	}

	// Create profile with review_days
	days30 := 30
	if err := SaveProfilesConfig(&ProfilesConfig{
		Profiles: []ProfileConfig{
			{Name: "test.trial", Stage: "trial", ReviewDays: &days30},
		},
	}); err != nil {
		t.Fatal(err)
	}

	// Add package without review_by (nil)
	pkg := PackageConfig{
		ID:          "formula:test-pkg",
		Name:        "test-pkg",
		Provider:    "brew",
		Profile:     "test.trial",
		InstalledAt: time.Now(),
		ReviewBy:    nil, // No review_by set
	}
	if err := AddPackage(pkg); err != nil {
		t.Fatal(err)
	}

	// GetOverduePackages should return this package
	overdue, err := GetOverduePackages()
	if err != nil {
		t.Fatal(err)
	}

	if len(overdue) != 1 {
		t.Errorf("Expected 1 overdue package (nil review_by), got %d", len(overdue))
	}
}

// TestGetOverduePackages_ProfileWithoutReviewDays tests that packages in profiles
// without review_days are NOT considered for review.
func TestGetOverduePackages_ProfileWithoutReviewDays(t *testing.T) {
	dir := t.TempDir()
	restore := setEnv("AL_HOME", dir)
	defer restore()

	if err := EnsureConfigDir(); err != nil {
		t.Fatal(err)
	}

	// Create profile WITHOUT review_days (like stable profile)
	if err := SaveProfilesConfig(&ProfilesConfig{
		Profiles: []ProfileConfig{
			{Name: "stable", Stage: "stable", ReviewDays: nil},
		},
	}); err != nil {
		t.Fatal(err)
	}

	// Add package with past review_by date
	pastDate := time.Date(2026, 2, 8, 0, 0, 0, 0, time.UTC)
	pkg := PackageConfig{
		ID:          "formula:test-pkg",
		Name:        "test-pkg",
		Provider:    "brew",
		Profile:     "stable",
		InstalledAt: time.Now(),
		ReviewBy:    &pastDate, // Past date, but profile has no review_days
	}
	if err := AddPackage(pkg); err != nil {
		t.Fatal(err)
	}

	// GetOverduePackages should return EMPTY (no review for this profile)
	overdue, err := GetOverduePackages()
	if err != nil {
		t.Fatal(err)
	}

	if len(overdue) != 0 {
		t.Errorf("Expected 0 overdue packages (profile has no review_days), got %d", len(overdue))
	}
}

// TestGetReviewDays_TrialProfileFromTemplate verifies that when a profile is created
// from the stable-trial template, the trial profile gets review_days.
func TestGetReviewDays_TrialProfileFromTemplate(t *testing.T) {
	dir := t.TempDir()
	restore := setEnv("AL_HOME", dir)
	defer restore()

	if err := EnsureConfigDir(); err != nil {
		t.Fatal(err)
	}

	// Get the stable-trial template
	template, err := GetTemplate("stable-trial")
	if err != nil {
		t.Fatal(err)
	}

	// Apply the template to create profiles
	profiles, err := ApplyTemplate(template, "myapp")
	if err != nil {
		t.Fatal(err)
	}

	// Save the profiles
	if err := SaveProfilesConfig(&ProfilesConfig{Profiles: profiles}); err != nil {
		t.Fatal(err)
	}

	// Check that myapp.trial has review_days
	days, hasReview, err := GetReviewDays("myapp.trial")
	if err != nil {
		t.Fatal(err)
	}

	if !hasReview {
		t.Error("Trial profile from stable-trial template should have review_days")
	}
	if days != 30 {
		t.Errorf("Expected review_days to be 30, got %d", days)
	}

	// Check that myapp (stable) does NOT have review_days
	days, hasReview, err = GetReviewDays("myapp")
	if err != nil {
		t.Fatal(err)
	}

	if hasReview {
		t.Error("Stable profile from stable-trial template should NOT have review_days")
	}
}
