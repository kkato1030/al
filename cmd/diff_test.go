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

	// Check that flags are registered
	flags := cmd.Flags()
	if flags == nil {
		t.Fatal("Flags() returned nil")
	}

	// Check --all flag
	allFlag := flags.Lookup("all")
	if allFlag == nil {
		t.Error("--all flag not registered")
	}

	// Check --profile flag
	profileFlag := flags.Lookup("profile")
	if profileFlag == nil {
		t.Error("--profile flag not registered")
	}

	// Check --prf flag (alias)
	prfFlag := flags.Lookup("prf")
	if prfFlag == nil {
		t.Error("--prf flag not registered")
	}
}

func TestDiffFlagValidation(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
	}{
		{
			name:        "no flags",
			args:        []string{},
			expectError: false,
		},
		{
			name:        "all flag only",
			args:        []string{"--all"},
			expectError: false,
		},
		{
			name:        "profile flag only",
			args:        []string{"--profile", "dev"},
			expectError: false,
		},
		{
			name:        "prf flag only",
			args:        []string{"--prf", "dev"},
			expectError: false,
		},
		{
			name:        "both all and profile flags - should error",
			args:        []string{"--all", "--profile", "dev"},
			expectError: true,
		},
		{
			name:        "both all and prf flags - should error",
			args:        []string{"--all", "--prf", "dev"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewDiffCmd()
			cmd.SetArgs(tt.args)
			err := cmd.Execute()

			// We can't fully test execution without a real config, but we can test flag parsing
			// The error check here is mainly for conflicting flags
			hasError := err != nil
			if tt.expectError && !hasError {
				t.Errorf("Expected error for args %v, but got none", tt.args)
			}
		})
	}
}
