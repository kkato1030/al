package config

import (
	"path/filepath"
	"testing"
	"time"
)

func TestAddPackage_GetOverduePackages(t *testing.T) {
	dir := t.TempDir()
	restore := setEnv("AL_HOME", dir)
	defer restore()

	if err := EnsureConfigDir(); err != nil {
		t.Fatal(err)
	}

	// Profile with review_days
	days30 := 30
	if err := SaveProfilesConfig(&ProfilesConfig{
		Profiles: []ProfileConfig{{Name: "reviewed", ReviewDays: &days30}},
	}); err != nil {
		t.Fatal(err)
	}

	// Add package without ReviewBy (should be overdue)
	pkg := PackageConfig{
		ID: "formula:ruby", Name: "ruby", Provider: "brew", Profile: "reviewed",
		InstalledAt: time.Now(),
	}
	if err := AddPackage(pkg); err != nil {
		t.Fatal(err)
	}

	overdue, err := GetOverduePackages()
	if err != nil {
		t.Fatal(err)
	}
	if len(overdue) != 1 || overdue[0].ID != "formula:ruby" {
		t.Errorf("GetOverduePackages() = %v", overdue)
	}
}

func TestGetOverduePackages_PastDate(t *testing.T) {
	dir := t.TempDir()
	restore := setEnv("AL_HOME", dir)
	defer restore()

	if err := EnsureConfigDir(); err != nil {
		t.Fatal(err)
	}

	// Profile with review_days
	days30 := 30
	if err := SaveProfilesConfig(&ProfilesConfig{
		Profiles: []ProfileConfig{{Name: "reviewed", ReviewDays: &days30}},
	}); err != nil {
		t.Fatal(err)
	}

	// Add package with ReviewBy set to past date (should be overdue)
	pastDate := time.Date(2026, 2, 8, 0, 0, 0, 0, time.UTC)
	pkg := PackageConfig{
		ID: "formula:nodejs", Name: "nodejs", Provider: "brew", Profile: "reviewed",
		InstalledAt: time.Now(),
		ReviewBy:    &pastDate,
	}
	if err := AddPackage(pkg); err != nil {
		t.Fatal(err)
	}

	overdue, err := GetOverduePackages()
	if err != nil {
		t.Fatal(err)
	}
	if len(overdue) != 1 || overdue[0].ID != "formula:nodejs" {
		t.Errorf("GetOverduePackages() with past ReviewBy = %v, expected 1 package", overdue)
	}
}

func TestGetOverduePackages_NilReviewBy(t *testing.T) {
	dir := t.TempDir()
	restore := setEnv("AL_HOME", dir)
	defer restore()

	if err := EnsureConfigDir(); err != nil {
		t.Fatal(err)
	}

	// Profile with review_days
	days30 := 30
	if err := SaveProfilesConfig(&ProfilesConfig{
		Profiles: []ProfileConfig{{Name: "reviewed", ReviewDays: &days30}},
	}); err != nil {
		t.Fatal(err)
	}

	// Add package with nil ReviewBy (should be overdue)
	pkg := PackageConfig{
		ID: "formula:python", Name: "python", Provider: "brew", Profile: "reviewed",
		InstalledAt: time.Now(),
		ReviewBy:    nil,
	}
	if err := AddPackage(pkg); err != nil {
		t.Fatal(err)
	}

	overdue, err := GetOverduePackages()
	if err != nil {
		t.Fatal(err)
	}
	if len(overdue) != 1 || overdue[0].ID != "formula:python" {
		t.Errorf("GetOverduePackages() with nil ReviewBy = %v, expected 1 package", overdue)
	}
}

func TestGetOverduePackages_ExplicitReviewByWithoutProfileReviewDays(t *testing.T) {
	dir := t.TempDir()
	restore := setEnv("AL_HOME", dir)
	defer restore()

	if err := EnsureConfigDir(); err != nil {
		t.Fatal(err)
	}

	// Profile WITHOUT review_days
	if err := SaveProfilesConfig(&ProfilesConfig{
		Profiles: []ProfileConfig{{Name: "stable"}},
	}); err != nil {
		t.Fatal(err)
	}

	// Add package with explicit past ReviewBy (should be overdue even though profile has no review_days)
	pastDate := time.Date(2026, 2, 8, 0, 0, 0, 0, time.UTC)
	pkg := PackageConfig{
		ID: "formula:golang", Name: "golang", Provider: "brew", Profile: "stable",
		InstalledAt: time.Now(),
		ReviewBy:    &pastDate,
	}
	if err := AddPackage(pkg); err != nil {
		t.Fatal(err)
	}

	overdue, err := GetOverduePackages()
	if err != nil {
		t.Fatal(err)
	}
	if len(overdue) != 1 || overdue[0].ID != "formula:golang" {
		t.Errorf("GetOverduePackages() with explicit past ReviewBy but no profile review_days = %v, expected 1 package", overdue)
	}
}

func TestGetOverduePackages_NoReviewByNoProfileReviewDays(t *testing.T) {
	dir := t.TempDir()
	restore := setEnv("AL_HOME", dir)
	defer restore()

	if err := EnsureConfigDir(); err != nil {
		t.Fatal(err)
	}

	// Profile WITHOUT review_days
	if err := SaveProfilesConfig(&ProfilesConfig{
		Profiles: []ProfileConfig{{Name: "stable"}},
	}); err != nil {
		t.Fatal(err)
	}

	// Add package without ReviewBy and profile has no review_days (should NOT be overdue)
	pkg := PackageConfig{
		ID: "formula:vim", Name: "vim", Provider: "brew", Profile: "stable",
		InstalledAt: time.Now(),
		ReviewBy:    nil,
	}
	if err := AddPackage(pkg); err != nil {
		t.Fatal(err)
	}

	overdue, err := GetOverduePackages()
	if err != nil {
		t.Fatal(err)
	}
	if len(overdue) != 0 {
		t.Errorf("GetOverduePackages() with no ReviewBy and no profile review_days = %v, expected 0 packages", overdue)
	}
}

func TestAddPackage_DuplicateError(t *testing.T) {
	dir := t.TempDir()
	restore := setEnv("AL_HOME", dir)
	defer restore()

	if err := EnsureConfigDir(); err != nil {
		t.Fatal(err)
	}

	pkg := PackageConfig{
		ID: "formula:ruby", Name: "ruby", Provider: "brew", Profile: "default",
		InstalledAt: time.Now(),
	}
	if err := AddPackage(pkg); err != nil {
		t.Fatal(err)
	}
	err := AddPackage(pkg)
	if err == nil {
		t.Error("AddPackage duplicate should return error")
	}
}

func TestAddOrUpdatePackage(t *testing.T) {
	dir := t.TempDir()
	restore := setEnv("AL_HOME", dir)
	defer restore()

	if err := EnsureConfigDir(); err != nil {
		t.Fatal(err)
	}

	installed := time.Now()
	pkg := PackageConfig{
		ID: "formula:foo", Name: "foo", Provider: "brew", Profile: "default",
		InstalledAt: installed,
	}
	if err := AddOrUpdatePackage(pkg); err != nil {
		t.Fatal(err)
	}

	// Update same package
	pkg.Description = "updated"
	if err := AddOrUpdatePackage(pkg); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadPackagesConfig()
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Packages) != 1 || cfg.Packages[0].Description != "updated" {
		t.Errorf("LoadPackagesConfig() = %+v", cfg.Packages)
	}
}

func TestSamePackageInOtherProfile(t *testing.T) {
	dir := t.TempDir()
	restore := setEnv("AL_HOME", dir)
	defer restore()

	if err := EnsureConfigDir(); err != nil {
		t.Fatal(err)
	}

	pkg1 := PackageConfig{ID: "formula:ruby", Name: "ruby", Provider: "brew", Profile: "default", InstalledAt: time.Now()}
	pkg2 := PackageConfig{ID: "formula:ruby", Name: "ruby", Provider: "brew", Profile: "work", InstalledAt: time.Now()}
	if err := AddPackage(pkg1); err != nil {
		t.Fatal(err)
	}
	if err := AddPackage(pkg2); err != nil {
		t.Fatal(err)
	}

	ok, err := SamePackageInOtherProfile("formula:ruby", "brew", "default")
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Error("SamePackageInOtherProfile should be true (work profile has same package)")
	}

	ok, err = SamePackageInOtherProfile("formula:ruby", "brew", "work")
	if err != nil || !ok {
		t.Errorf("SamePackageInOtherProfile(work) = %v, %v", ok, err)
	}

	ok, err = SamePackageInOtherProfile("formula:nonexistent", "brew", "default")
	if err != nil || ok {
		t.Errorf("nonexistent: ok=%v err=%v", ok, err)
	}
}

func TestRemovePackage(t *testing.T) {
	dir := t.TempDir()
	restore := setEnv("AL_HOME", dir)
	defer restore()

	if err := EnsureConfigDir(); err != nil {
		t.Fatal(err)
	}

	pkg := PackageConfig{ID: "formula:bar", Name: "bar", Provider: "brew", Profile: "default", InstalledAt: time.Now()}
	if err := AddPackage(pkg); err != nil {
		t.Fatal(err)
	}

	if err := RemovePackage("formula:bar", "brew", "default"); err != nil {
		t.Fatal(err)
	}

	cfg, _ := LoadPackagesConfig()
	if len(cfg.Packages) != 0 {
		t.Errorf("packages after remove: %v", cfg.Packages)
	}

	err := RemovePackage("formula:nonexistent", "brew", "default")
	if err == nil {
		t.Error("RemovePackage nonexistent should return error")
	}
}

func TestSetPackageReviewBy(t *testing.T) {
	dir := t.TempDir()
	restore := setEnv("AL_HOME", dir)
	defer restore()

	if err := EnsureConfigDir(); err != nil {
		t.Fatal(err)
	}

	pkg := PackageConfig{ID: "formula:baz", Name: "baz", Provider: "brew", Profile: "default", InstalledAt: time.Now()}
	if err := AddPackage(pkg); err != nil {
		t.Fatal(err)
	}

	reviewBy := time.Now().Add(7 * 24 * time.Hour)
	if err := SetPackageReviewBy("formula:baz", "brew", "default", reviewBy); err != nil {
		t.Fatal(err)
	}

	cfg, _ := LoadPackagesConfig()
	if len(cfg.Packages) != 1 || cfg.Packages[0].ReviewBy == nil {
		t.Errorf("ReviewBy not set: %+v", cfg.Packages[0])
	}

	err := SetPackageReviewBy("formula:nonexistent", "brew", "default", reviewBy)
	if err == nil {
		t.Error("SetPackageReviewBy nonexistent should return error")
	}
}

func TestLoadPackagesConfig_NoFile(t *testing.T) {
	dir := t.TempDir()
	restore := setEnv("AL_HOME", dir)
	defer restore()

	// No packages.json
	cfg, err := LoadPackagesConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg == nil || len(cfg.Packages) != 0 {
		t.Errorf("LoadPackagesConfig() = %v", cfg)
	}
}

func TestGetPackagesConfigPath(t *testing.T) {
	dir := t.TempDir()
	restore := setEnv("AL_HOME", dir)
	defer restore()

	got, err := GetPackagesConfigPath()
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(dir, "packages.json")
	if got != want {
		t.Errorf("GetPackagesConfigPath() = %q, want %q", got, want)
	}
}
