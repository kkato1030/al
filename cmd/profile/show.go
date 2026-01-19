package profile

import (
	"encoding/json"
	"fmt"

	"github.com/kkato1030/al/internal/config"
	"github.com/spf13/cobra"
)

// NewProfileShowCmd creates the profile show command
func NewProfileShowCmd() *cobra.Command {
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "show [profile-name]",
		Short: "Show profile details",
		Long:  "Show detailed information about a specific profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProfileShow(args[0], outputFormat)
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "output", "o", "json", "Output format (json)")

	return cmd
}

func runProfileShow(profileName, outputFormat string) error {
	profile, err := config.GetProfile(profileName)
	if err != nil {
		return fmt.Errorf("error loading profile: %w", err)
	}

	if profile == nil {
		return fmt.Errorf("profile '%s' not found", profileName)
	}

	switch outputFormat {
	case "json":
		return outputJSON(profile)
	default:
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}
}

func outputJSON(profile *config.ProfileConfig) error {
	data, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling profile to JSON: %w", err)
	}

	fmt.Println(string(data))
	return nil
}
