package shell

import (
	"github.com/kkato1030/al/internal/config"
	"github.com/kkato1030/al/internal/output"
	"github.com/kkato1030/al/internal/ui"
	"github.com/spf13/cobra"
)

// NewEnableCmd creates the shell enable command
func NewEnableCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "enable <package-name>",
		Short: "Enable package shell snippet for al activate",
		Args:  cobra.ExactArgs(1),
		RunE:  runEnable,
	}
	return cmd
}

func runEnable(cmd *cobra.Command, args []string) error {
	return setEnabled(args[0], true)
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
	output.Success("%s shell snippet for %s", verb, pkg.Name)
	return nil
}
