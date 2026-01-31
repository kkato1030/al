package link

import (
	"fmt"

	"github.com/kkato1030/al/internal/config"
	"github.com/kkato1030/al/internal/ui"
	"github.com/spf13/cobra"
)

// NewAddCmd creates the package link add command
func NewAddCmd() *cobra.Command {
	var path string
	cmd := &cobra.Command{
		Use:   "add <package_name> --path <path>",
		Short: "Add link.d entry for a package",
		Long:  "Add a path to link.d under the package name. Link name = package name. Resolves package by name (interactive if multiple matches).",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if path == "" {
				return fmt.Errorf("--path is required")
			}
			return runAdd(args[0], path)
		},
	}
	cmd.Flags().StringVar(&path, "path", "", "Path (symlink location)")
	cmd.MarkFlagRequired("path")
	return cmd
}

func runAdd(packageName, userPath string) error {
	pkg, err := ui.ResolvePackageByName(packageName)
	if err != nil {
		return fmt.Errorf("resolving package: %w", err)
	}
	linkType, err := config.DetectLinkType(userPath)
	if err != nil {
		return err
	}
	// Link name = package name (same)
	entry, err := config.AddLink(pkg.Name, userPath, linkType, pkg.ID, pkg.Provider)
	if err != nil {
		return err
	}
	fmt.Printf("Added link %s -> %s (type: %s)\n", entry.Name, entry.Manifest.UserPath, linkType)
	return nil
}
