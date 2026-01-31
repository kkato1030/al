package link

import (
	"fmt"

	"github.com/kkato1030/al/internal/config"
	"github.com/kkato1030/al/internal/ui"
	"github.com/spf13/cobra"
)

// NewRemoveCmd creates the link remove command
func NewRemoveCmd() *cobra.Command {
	var path string
	var pkgName string
	var purge bool
	cmd := &cobra.Command{
		Use:   "remove --path <path> [--package <pkg>] [--purge]",
		Short: "Remove a link from link.d",
		Long:  "Remove the symlink and copy content back to the path (default). Use --purge to delete the link.d content without copy-back.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if path == "" {
				return fmt.Errorf("--path is required")
			}
			return runRemove(path, pkgName, purge)
		},
	}
	cmd.Flags().StringVar(&path, "path", "", "Path (symlink location) to remove")
	cmd.Flags().StringVar(&pkgName, "package", "", "Filter by package name (optional)")
	cmd.Flags().BoolVar(&purge, "purge", false, "Delete link.d content without copy-back")
	return cmd
}

func runRemove(userPath, pkgName string, purge bool) error {
	var packageID, packageProvider string
	if pkgName != "" {
		pkg, err := ui.ResolvePackageByName(pkgName)
		if err != nil {
			return fmt.Errorf("resolving package: %w", err)
		}
		packageID = pkg.ID
		packageProvider = pkg.Provider
	}
	entry, entryDir, err := config.FindLinkByUserPath(userPath, packageID, packageProvider)
	if err != nil {
		return err
	}
	if entry == nil {
		return fmt.Errorf("link not found for path %s", userPath)
	}
	if err := config.RemoveLink(entry, entryDir, purge); err != nil {
		return err
	}
	verb := "Removed"
	if purge {
		verb = "Purged"
	}
	fmt.Printf("%s link %s\n", verb, entry.Manifest.UserPath)
	return nil
}
