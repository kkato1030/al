package config

import (
	"fmt"
	"sort"

	"github.com/kkato1030/al/internal/config"
	"github.com/spf13/cobra"
)

// NewConfigAliasCmd creates the config alias command
func NewConfigAliasCmd() *cobra.Command {
	aliasCmd := &cobra.Command{
		Use:   "alias",
		Short: "Manage command aliases",
		Long:  "Manage command aliases",
	}

	aliasCmd.AddCommand(NewConfigAliasListCmd())

	return aliasCmd
}

// NewConfigAliasListCmd creates the config alias list command
func NewConfigAliasListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all command aliases",
		Long:  "List all available command aliases",
		RunE: func(cmd *cobra.Command, args []string) error {
			aliases := config.GetDefaultAliases()
			if len(aliases) == 0 {
				fmt.Println("No aliases configured")
				return nil
			}

			fmt.Println("Available aliases:")

			// Sort alias names for consistent output
			aliasNames := make([]string, 0, len(aliases))
			for name := range aliases {
				aliasNames = append(aliasNames, name)
			}
			sort.Strings(aliasNames)

			for _, name := range aliasNames {
				fmt.Printf("  %-10s %s\n", name, aliases[name])
			}

			return nil
		},
	}
}
