package profile

import (
	"fmt"

	"github.com/kkato1030/al/internal/config"
	"github.com/spf13/cobra"
)

// NewProfileListCmd creates the profile list command
func NewProfileListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all profiles",
		Long:  "List all configured profiles",
		RunE:  runProfileList,
	}
}

func runProfileList(cmd *cobra.Command, args []string) error {
	profilesConfig, err := config.LoadProfilesConfig()
	if err != nil {
		return fmt.Errorf("error loading profiles config: %w", err)
	}

	if len(profilesConfig.Profiles) == 0 {
		fmt.Println("No profiles configured")
		return nil
	}

	fmt.Println("Configured profiles:")
	for _, p := range profilesConfig.Profiles {
		if p.Description != "" {
			fmt.Printf("  - %s (description: %s)\n", p.Name, p.Description)
		} else {
			fmt.Printf("  - %s\n", p.Name)
		}
	}

	return nil
}
