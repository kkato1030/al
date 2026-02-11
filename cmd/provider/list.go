package provider

import (
	"fmt"
	"strings"

	"github.com/kkato1030/al/internal/config"
	"github.com/spf13/cobra"
)

// NewProviderListCmd creates the provider list command
func NewProviderListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all providers",
		Long:  "List all installed package manager providers",
		RunE:  runProviderList,
	}
}

func runProviderList(cmd *cobra.Command, args []string) error {
	providersCfg, err := config.LoadProvidersConfig()
	if err != nil {
		return fmt.Errorf("error loading providers config: %w", err)
	}

	if len(providersCfg.Providers) == 0 {
		fmt.Println("No providers installed")
		return nil
	}

	fmt.Println("Installed providers:")
	for _, p := range providersCfg.Providers {
		fmt.Printf("  - %s", p.Name)
		if p.Version != "" {
			fmt.Printf(" (version: %s)", p.Version)
		}
		if !p.InstalledAt.IsZero() {
			fmt.Printf(" (installed at: %s)", p.InstalledAt.Format("2006-01-02 15:04:05"))
		}
		deps := p.DependsOn
		if deps == nil {
			deps = config.GetDefaultProviderDependencies(p.Name)
		}
		if len(deps) > 0 {
			fmt.Printf(" (depends_on: %s)", strings.Join(deps, ", "))
		}
		fmt.Println()
	}

	return nil
}
