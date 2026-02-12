package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kkato1030/al/internal/config"
)

func TestDoctorCommand_NoInit(t *testing.T) {
	// Set up a temporary config directory that doesn't exist
	tmpDir := t.TempDir()
	nonExistentDir := filepath.Join(tmpDir, "nonexistent")
	origAlHome := os.Getenv("AL_HOME")
	os.Setenv("AL_HOME", nonExistentDir)
	defer os.Setenv("AL_HOME", origAlHome)

	// Don't initialize - should report config directory doesn't exist
	results := checkConfigDir()

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	if results[0].Status != StatusError {
		t.Errorf("Expected ERROR status, got %v", results[0].Status)
	}
}

func TestDoctorCommand_WithInit(t *testing.T) {
	// Set up a temporary config directory
	tmpDir := t.TempDir()
	origAlHome := os.Getenv("AL_HOME")
	os.Setenv("AL_HOME", tmpDir)
	defer os.Setenv("AL_HOME", origAlHome)

	// Create config directory structure
	if err := config.EnsureConfigDir(); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	// Check that it's now OK
	results := checkConfigDir()

	foundOK := false
	for _, r := range results {
		if r.Status == StatusOK {
			foundOK = true
			break
		}
	}

	if !foundOK {
		t.Error("Expected at least one OK status after init")
	}
}

func TestDoctorCommand_ConfigFiles(t *testing.T) {
	// Set up a temporary config directory
	tmpDir := t.TempDir()
	origAlHome := os.Getenv("AL_HOME")
	os.Setenv("AL_HOME", tmpDir)
	defer os.Setenv("AL_HOME", origAlHome)

	// Create config directory
	if err := config.EnsureConfigDir(); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	// Create basic config files
	cfg := &config.AppConfig{
		DefaultProvider: "brew",
		DefaultProfile:  "core.trial",
		DefaultStage:    "trial",
	}
	if err := config.SaveAppConfig(cfg); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Create profiles
	profilesCfg := &config.ProfilesConfig{
		Profiles: []config.ProfileConfig{
			{Name: "core.trial", Stage: "trial"},
		},
	}
	if err := config.SaveProfilesConfig(profilesCfg); err != nil {
		t.Fatalf("Failed to save profiles: %v", err)
	}

	// Check config files
	results := checkConfigFiles()

	okCount := 0
	for _, r := range results {
		if r.Status == StatusOK {
			okCount++
		}
	}

	if okCount == 0 {
		t.Error("Expected some OK statuses for valid config files")
	}
}

func TestDoctorCommand_BrokenLink(t *testing.T) {
	// Set up a temporary config directory
	tmpDir := t.TempDir()
	origAlHome := os.Getenv("AL_HOME")
	os.Setenv("AL_HOME", tmpDir)
	defer os.Setenv("AL_HOME", origAlHome)

	// Create config directory
	if err := config.EnsureConfigDir(); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	// Create a link that points to a non-existent file
	linkDir, _ := config.GetLinkDir()
	entryDir := filepath.Join(linkDir, "testlink")
	if err := os.MkdirAll(entryDir, 0755); err != nil {
		t.Fatalf("Failed to create link entry dir: %v", err)
	}

	// Create manifest
	manifestData := []byte(`{"user_path":"/nonexistent/path","type":"file"}`)
	if err := os.WriteFile(filepath.Join(entryDir, ".manifest.json"), manifestData, 0644); err != nil {
		t.Fatalf("Failed to write manifest: %v", err)
	}

	// Create content file
	contentPath := filepath.Join(entryDir, "content")
	if err := os.WriteFile(contentPath, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to write content: %v", err)
	}

	// Create symlink (it will point to content, but manifest says it should be at /nonexistent/path)
	// For testing, we just check if the symlink at the manifest path exists

	// Check links - should report that the symlink doesn't exist at user path
	results := checkLinks()

	foundWarn := false
	for _, r := range results {
		if r.Status == StatusWarn && r.Message != "" {
			foundWarn = true
			break
		}
	}

	if !foundWarn {
		t.Error("Expected a warning for missing/broken symlink")
	}
}

func TestDoctorCommand_InvalidProfileReference(t *testing.T) {
	// Set up a temporary config directory
	tmpDir := t.TempDir()
	origAlHome := os.Getenv("AL_HOME")
	os.Setenv("AL_HOME", tmpDir)
	defer os.Setenv("AL_HOME", origAlHome)

	// Create config directory
	if err := config.EnsureConfigDir(); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	// Create profiles config with one profile
	profilesCfg := &config.ProfilesConfig{
		Profiles: []config.ProfileConfig{
			{Name: "core.trial", Stage: "trial"},
		},
	}
	if err := config.SaveProfilesConfig(profilesCfg); err != nil {
		t.Fatalf("Failed to save profiles: %v", err)
	}

	// Create packages config with reference to non-existent profile
	packagesCfg := &config.PackagesConfig{
		Packages: []config.PackageConfig{
			{
				ID:       "test",
				Name:     "test-package",
				Provider: "brew",
				Profile:  "nonexistent",
			},
		},
	}
	if err := config.SavePackagesConfig(packagesCfg); err != nil {
		t.Fatalf("Failed to save packages: %v", err)
	}

	// Check for invalid references
	results := checkInvalidProfileReferences()

	foundWarn := false
	for _, r := range results {
		if r.Status == StatusWarn && r.Message != "" {
			foundWarn = true
			break
		}
	}

	if !foundWarn {
		t.Error("Expected a warning for invalid profile reference")
	}
}

func TestPrintResults(t *testing.T) {
	// Just test that it doesn't panic
	results := []CheckResult{
		{Status: StatusOK, Message: "test ok"},
		{Status: StatusWarn, Message: "test warn"},
		{Status: StatusError, Message: "test error"},
	}

	// This will print to stdout but shouldn't panic
	printResults(results)
}
