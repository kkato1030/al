package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kkato1030/al/internal/config"
)

func TestRunInit_Basic(t *testing.T) {
	// Set up a temporary config directory
	tmpDir := t.TempDir()
	origAlHome := os.Getenv("AL_HOME")
	os.Setenv("AL_HOME", tmpDir)
	defer os.Setenv("AL_HOME", origAlHome)

	// Run init
	if err := runInit(nil, nil); err != nil {
		t.Fatalf("Failed to run init: %v", err)
	}

	// Verify config directory was created
	if _, err := os.Stat(tmpDir); os.IsNotExist(err) {
		t.Error("Config directory was not created")
	}

	// Verify profiles were created
	profilesConfig, err := config.LoadProfilesConfig()
	if err != nil {
		t.Fatalf("Failed to load profiles config: %v", err)
	}

	if len(profilesConfig.Profiles) != 2 {
		t.Errorf("Expected 2 profiles, got %d", len(profilesConfig.Profiles))
	}

	// Check for core and core.trial profiles
	var foundCore, foundCoreTrial bool
	for _, p := range profilesConfig.Profiles {
		if p.Name == "core" && p.Stage == "stable" {
			foundCore = true
		}
		if p.Name == "core.trial" && p.Stage == "trial" {
			foundCoreTrial = true
		}
	}

	if !foundCore {
		t.Error("core profile was not created")
	}
	if !foundCoreTrial {
		t.Error("core.trial profile was not created")
	}

	// Verify app config
	appConfig, err := config.LoadAppConfig()
	if err != nil {
		t.Fatalf("Failed to load app config: %v", err)
	}

	if appConfig.DefaultProfile != "core.trial" {
		t.Errorf("Expected default_profile to be 'core.trial', got '%s'", appConfig.DefaultProfile)
	}

	if appConfig.DefaultProvider != "brew" {
		t.Errorf("Expected default_provider to be 'brew', got '%s'", appConfig.DefaultProvider)
	}

	if appConfig.DefaultStage != "trial" {
		t.Errorf("Expected default_stage to be 'trial', got '%s'", appConfig.DefaultStage)
	}

	// Verify .gitignore was created
	gitignorePath := filepath.Join(tmpDir, ".gitignore")
	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		t.Error(".gitignore was not created")
	}
}

func TestRunInit_IdempotentConfigDir(t *testing.T) {
	// Set up a temporary config directory
	tmpDir := t.TempDir()
	origAlHome := os.Getenv("AL_HOME")
	os.Setenv("AL_HOME", tmpDir)
	defer os.Setenv("AL_HOME", origAlHome)

	// Create the config directory first
	if err := config.EnsureConfigDir(); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	// Run init - should not fail even if directory exists
	if err := runInit(nil, nil); err != nil {
		t.Fatalf("Failed to run init on existing directory: %v", err)
	}
}
