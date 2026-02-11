package shell

import (
	"fmt"

	"github.com/kkato1030/al/internal/config"
	"github.com/kkato1030/al/internal/ui"
	"github.com/spf13/cobra"
)

// NewRemoveCmd creates the shell remove command
func NewRemoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove <package-name>",
		Short: "Remove package shell.d snippet",
		Long:  "Remove the shell.d directory for the package (snippet and manifest).",
		Args:  cobra.ExactArgs(1),
		RunE:  runRemove,
	}
	return cmd
}

func runRemove(cmd *cobra.Command, args []string) error {
	pkg, err := ui.ResolvePackageByName(args[0])
	if err != nil {
		return err
	}
	if err := config.RemoveShellPackageDir(pkg.ID, pkg.Provider); err != nil {
		return fmt.Errorf("removing shell.d: %w", err)
	}
	fmt.Printf("Removed shell snippet for %s (provider: %s)\n", pkg.Name, pkg.Provider)
	return nil
}
