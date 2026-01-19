package packagecmd

import (
	"encoding/json"
	"fmt"

	"github.com/kkato1030/al/internal/config"
	"github.com/spf13/cobra"
)

// NewPackageShowCmd creates the package show command
func NewPackageShowCmd() *cobra.Command {
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "show <package-name>",
		Short: "Show package details",
		Long:  "Show detailed information about a specific package",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPackageShow(args[0], outputFormat)
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "output", "o", "json", "Output format (json)")

	return cmd
}

func runPackageShow(packageName, outputFormat string) error {
	packagesConfig, err := config.LoadPackagesConfig()
	if err != nil {
		return fmt.Errorf("error loading packages config: %w", err)
	}

	// Find all packages with the given name
	var matchingPackages []config.PackageConfig
	for _, pkg := range packagesConfig.Packages {
		if pkg.Name == packageName {
			matchingPackages = append(matchingPackages, pkg)
		}
	}

	if len(matchingPackages) == 0 {
		return fmt.Errorf("package '%s' not found", packageName)
	}

	switch outputFormat {
	case "json":
		return outputJSON(matchingPackages)
	default:
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}
}

func outputJSON(packages []config.PackageConfig) error {
	// If only one package, output it directly; otherwise output as array
	var data []byte
	var err error
	if len(packages) == 1 {
		data, err = json.MarshalIndent(packages[0], "", "  ")
	} else {
		data, err = json.MarshalIndent(packages, "", "  ")
	}
	if err != nil {
		return fmt.Errorf("error marshaling package to JSON: %w", err)
	}

	fmt.Println(string(data))
	return nil
}
