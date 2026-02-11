package packagecmd

import (
	"os"
	"testing"

	"github.com/kkato1030/al/internal/config"
)

// TestFindProfileWithFallback tests the fallback logic for finding profiles
func TestFindProfileWithFallback(t *testing.T) {
	// Create a temporary directory for test config
	tmpDir, err := os.MkdirTemp("", "al-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Set AL_HOME to the temp directory
	originalALHome := os.Getenv("AL_HOME")
	os.Setenv("AL_HOME", tmpDir)
	defer os.Setenv("AL_HOME", originalALHome)

	// Create test profiles
	profilesConfig := &config.ProfilesConfig{
		Profiles: []config.ProfileConfig{
			{Name: "core"},
			{Name: "core.stable", Stage: "stable"},
			{Name: "work.trial", Stage: "trial"},
			{Name: "private.stable", Stage: "stable"},
		},
	}

	// Save the profiles config
	if err := config.SaveProfilesConfig(profilesConfig); err != nil {
		t.Fatalf("Failed to save profiles config: %v", err)
	}

	tests := []struct {
		name            string
		fullProfileName string
		stageFlag       string
		wantProfileName string
		wantNil         bool
	}{
		{
			name:            "exact match exists",
			fullProfileName: "core.stable",
			stageFlag:       "stable",
			wantProfileName: "core.stable",
			wantNil:         false,
		},
		{
			name:            "profile without stage exists",
			fullProfileName: "core.trial",
			stageFlag:       "trial",
			wantProfileName: "core",
			wantNil:         false,
		},
		{
			name:            "fallback to different stage",
			fullProfileName: "work.stable",
			stageFlag:       "stable",
			wantProfileName: "work.trial",
			wantNil:         false,
		},
		{
			name:            "fallback to stable when trial not found",
			fullProfileName: "private.trial",
			stageFlag:       "trial",
			wantProfileName: "private.stable",
			wantNil:         false,
		},
		{
			name:            "profile not found at all",
			fullProfileName: "nonexistent.trial",
			stageFlag:       "trial",
			wantNil:         true,
		},
		{
			name:            "no stage flag, no fallback",
			fullProfileName: "nonexistent",
			stageFlag:       "",
			wantNil:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := findProfileWithFallback(tt.fullProfileName, tt.stageFlag)
			if err != nil {
				t.Fatalf("findProfileWithFallback() error = %v", err)
			}

			if tt.wantNil {
				if result != nil {
					t.Errorf("findProfileWithFallback() = %v, want nil", result)
				}
			} else {
				if result == nil {
					t.Errorf("findProfileWithFallback() = nil, want %s", tt.wantProfileName)
				} else if result.Name != tt.wantProfileName {
					t.Errorf("findProfileWithFallback() = %s, want %s", result.Name, tt.wantProfileName)
				}
			}
		})
	}
}
