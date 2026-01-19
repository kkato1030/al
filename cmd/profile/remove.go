package profile

import (
	"fmt"

	"github.com/kkato1030/al/internal/config"
	"github.com/spf13/cobra"
)

// NewProfileRemoveCmd creates the profile remove command
func NewProfileRemoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove [profile-name]",
		Short: "Remove a profile",
		Long:  "Remove a profile from the configuration",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProfileRemove(args[0])
		},
	}

	return cmd
}

func runProfileRemove(profileName string) error {
	// Check if profile exists
	profile, err := config.GetProfile(profileName)
	if err != nil {
		return fmt.Errorf("error loading profile: %w", err)
	}

	if profile == nil {
		return fmt.Errorf("profile '%s' not found", profileName)
	}

	// Remove the profile
	if err := config.RemoveProfile(profileName); err != nil {
		return fmt.Errorf("error removing profile: %w", err)
	}

	fmt.Printf("Profile '%s' has been successfully removed\n", profileName)
	return nil
}
