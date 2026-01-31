package shell

import (
	"fmt"

	"github.com/kkato1030/al/internal/config"
	"github.com/kkato1030/al/internal/ui"
	"github.com/spf13/cobra"
)

// NewDisableCmd creates the shell disable command
func NewDisableCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "disable <package-name>",
		Short: "Disable package shell snippet for al activate (file is kept)",
		Args:  cobra.ExactArgs(1),
		RunE:  runDisable,
	}
	return cmd
}

func runDisable(cmd *cobra.Command, args []string) error {
	return setEnabled(args[0], false)
}

func setEnabled(packageName string, enabled bool) error {
	pkg, err := ui.ResolvePackageByName(packageName)
	if err != nil {
		return err
	}
	pkgDir, err := config.GetShellPackageDir(pkg.ID, pkg.Provider)
	if err != nil {
		return err
	}
	manifest, err := config.LoadShellManifest(pkgDir)
	if err != nil {
		return err
	}
	manifest.Enabled = enabled
	if err := config.SaveShellManifest(pkgDir, manifest); err != nil {
		return err
	}
	verb := "Disabled"
	if enabled {
		verb = "Enabled"
	}
	fmt.Printf("%s shell snippet for %s (provider: %s, profile: %s)\n", verb, pkg.Name, pkg.Provider, pkg.Profile)
	return nil
}
