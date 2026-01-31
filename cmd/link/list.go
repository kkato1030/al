package link

import (
	"fmt"

	"github.com/kkato1030/al/internal/config"
	"github.com/kkato1030/al/internal/ui"
	"github.com/spf13/cobra"
)

// NewListCmd creates the link list command
func NewListCmd() *cobra.Command {
	var pkgName string
	cmd := &cobra.Command{
		Use:   "list [--package <pkg>]",
		Short: "List link.d entries",
		Long:  "List managed links. Use --package to filter by package name.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(pkgName)
		},
	}
	cmd.Flags().StringVar(&pkgName, "package", "", "Filter by package name")
	return cmd
}

func runList(pkgName string) error {
	var packageID, packageProvider string
	if pkgName != "" {
		pkg, err := ui.ResolvePackageByName(pkgName)
		if err != nil {
			return fmt.Errorf("resolving package: %w", err)
		}
		packageID = pkg.ID
		packageProvider = pkg.Provider
	}
	links, err := config.ListLinks(packageID, packageProvider)
	if err != nil {
		return err
	}
	if len(links) == 0 {
		fmt.Println("(no links)")
		return nil
	}
	for _, l := range links {
		pkgInfo := ""
		if l.Manifest.PackageID != "" {
			pkgInfo = fmt.Sprintf(" [package: %s/%s]", l.Manifest.PackageID, l.Manifest.PackageProvider)
		}
		fmt.Printf("%s -> %s (%s)%s\n", l.Name, l.Manifest.UserPath, l.Manifest.Type, pkgInfo)
	}
	return nil
}
