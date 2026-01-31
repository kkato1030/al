package link

import (
	"fmt"

	"github.com/kkato1030/al/internal/config"
	"github.com/kkato1030/al/internal/ui"
	"github.com/spf13/cobra"
)

// NewAddCmd creates the link add command
func NewAddCmd() *cobra.Command {
	var path string
	var pkgName string
	cmd := &cobra.Command{
		Use:   "add [--path] <path> [--package <pkg>]",
		Short: "Add a path to link.d",
		Long:  "Add a file or directory to link.d. The path becomes a symlink. Use --package to associate with a package (by name); type is inferred (existing path: by stat; non-existing: trailing / = dir, else file).",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			userPath := path
			if len(args) > 0 {
				userPath = args[0]
			}
			if userPath == "" {
				return fmt.Errorf("path is required (positional or --path)")
			}
			return runAdd(userPath, pkgName)
		},
	}
	cmd.Flags().StringVar(&path, "path", "", "Path to add (same as positional arg)")
	cmd.Flags().StringVar(&pkgName, "package", "", "Package name to associate (optional)")
	return cmd
}

func runAdd(userPath, pkgName string) error {
	var packageID, packageProvider string
	if pkgName != "" {
		pkg, err := ui.ResolvePackageByName(pkgName)
		if err != nil {
			return fmt.Errorf("resolving package: %w", err)
		}
		packageID = pkg.ID
		packageProvider = pkg.Provider
	}
	linkType, err := config.DetectLinkType(userPath)
	if err != nil {
		return err
	}
	entry, err := config.AddLink(userPath, linkType, packageID, packageProvider)
	if err != nil {
		return err
	}
	fmt.Printf("Added link %s -> %s (type: %s)\n", entry.Manifest.UserPath, entry.ID, linkType)
	return nil
}
