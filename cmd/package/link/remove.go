package link

import (
	"fmt"
	"path/filepath"

	"github.com/kkato1030/al/internal/config"
	"github.com/kkato1030/al/internal/ui"
	"github.com/spf13/cobra"
)

// NewRemoveCmd creates the package link remove command
func NewRemoveCmd() *cobra.Command {
	var purge bool
	cmd := &cobra.Command{
		Use:   "remove <package_name> [--purge]",
		Short: "Remove link.d entry for a package",
		Long:  "Remove the link associated with the package (one link per package). Copy-back by default; use --purge to delete without copy-back.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRemove(args[0], purge)
		},
	}
	cmd.Flags().BoolVar(&purge, "purge", false, "Delete link.d content without copy-back")
	return cmd
}

func runRemove(packageName string, purge bool) error {
	pkg, err := ui.ResolvePackageByName(packageName)
	if err != nil {
		return fmt.Errorf("resolving package: %w", err)
	}
	links, err := config.ListLinks(pkg.ID, pkg.Provider)
	if err != nil {
		return err
	}
	if len(links) == 0 {
		return fmt.Errorf("no link for package %s", pkg.Name)
	}
	entry := &links[0]
	linkDir, err := config.GetLinkDir()
	if err != nil {
		return err
	}
	entryDir := filepath.Join(linkDir, entry.Name)
	if err := config.RemoveLink(entry, entryDir, purge); err != nil {
		return err
	}
	verb := "Removed"
	if purge {
		verb = "Purged"
	}
	fmt.Printf("%s link %s\n", verb, entry.Name)
	return nil
}
