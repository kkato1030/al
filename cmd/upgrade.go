package cmd

import (
	"fmt"
	"strings"

	providercmd "github.com/kkato1030/al/cmd/provider"
	packagecmd "github.com/kkato1030/al/cmd/package"
	"github.com/spf13/cobra"
)

// NewUpgradeCmd creates the upgrade command
func NewUpgradeCmd() *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade all providers and packages",
		Long:  "Upgrade all providers and packages. This is equivalent to running 'al provider upgrade' followed by 'al package upgrade'.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpgrade(yes)
		},
	}

	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}

func runUpgrade(yes bool) error {
	// Ask for confirmation
	if !yes {
		fmt.Println("This will upgrade all providers and packages.")
		fmt.Println("This is equivalent to:")
		fmt.Println("  1. al provider upgrade")
		fmt.Println("  2. al package upgrade")
		fmt.Print("\nDo you want to continue? [y/N]: ")
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			fmt.Println("Upgrade cancelled.")
			return nil
		}
	}

	// Upgrade all providers
	fmt.Println()
	if err := providercmd.RunProviderUpgradeAll(true); err != nil {
		fmt.Printf("\nError upgrading providers: %v\n", err)
		// Continue to package upgrade even if provider upgrade fails
	}

	// Upgrade all packages
	fmt.Println()
	if err := packagecmd.RunPackageUpgradeAll(true); err != nil {
		return fmt.Errorf("error upgrading packages: %w", err)
	}

	fmt.Println("\n✓ All upgrades completed")
	return nil
}
