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
		Use:   "add <name> --path <path> [--package <pkg>]",
		Short: "Add a path to link.d",
		Long:  "Add a file or directory to link.d under the given name. The path becomes a symlink to link.d/<name>. Use --package to associate with a package; type is inferred (existing path: by stat; non-existing: trailing / = dir, else file).",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			if path == "" {
				return fmt.Errorf("--path is required")
			}
			return runAdd(name, path, pkgName)
		},
	}
	cmd.Flags().StringVar(&path, "path", "", "Path to add (symlink location)")
	cmd.MarkFlagRequired("path")
	cmd.Flags().StringVar(&pkgName, "package", "", "Package name to associate (optional)")
	return cmd
}

func runAdd(name, userPath, pkgName string) error {
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
	entry, err := config.AddLink(name, userPath, linkType, packageID, packageProvider)
	if err != nil {
		return err
	}
	fmt.Printf("Added link %s -> %s (type: %s)\n", entry.Name, entry.Manifest.UserPath, linkType)
	return nil
}
