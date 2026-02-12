package cmd

import (
	"testing"
)

func TestNewDiffCmd(t *testing.T) {
	cmd := NewDiffCmd()
	if cmd == nil {
		t.Fatal("NewDiffCmd returned nil")
	}

	if cmd.Use != "diff" {
		t.Errorf("Use = %q, want %q", cmd.Use, "diff")
	}

	if cmd.Short == "" {
		t.Error("Short description is empty")
	}

	// Check that no profile-specific flags are registered (we compare against all profiles now)
	flags := cmd.Flags()
	if flags == nil {
		t.Fatal("Flags() returned nil")
	}

	// These flags should no longer exist
	if flags.Lookup("all") != nil {
		t.Error("--all flag should not be registered")
	}
	if flags.Lookup("profile") != nil {
		t.Error("--profile flag should not be registered")
	}
	if flags.Lookup("prf") != nil {
		t.Error("--prf flag should not be registered")
	}
}

func TestDiffExecution(t *testing.T) {
	// Simple test to ensure the command can be created and executed
	// without crashing (it will fail due to missing config, which is expected)
	cmd := NewDiffCmd()
	cmd.SetArgs([]string{})

	// Execute will fail due to missing config, but should not panic
	_ = cmd.Execute()
}
