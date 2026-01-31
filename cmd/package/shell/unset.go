package shell

import (
	"fmt"

	"github.com/kkato1030/al/internal/config"
	"github.com/kkato1030/al/internal/ui"
	"github.com/spf13/cobra"
)

// NewUnsetCmd creates the shell unset command
func NewUnsetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unset <package-name>",
		Short: "Remove package shell.d snippet",
		Long:  "Remove the shell.d directory for the package (snippet and manifest).",
		Args:  cobra.ExactArgs(1),
		RunE:  runUnset,
	}
	return cmd
}

func runUnset(cmd *cobra.Command, args []string) error {
	pkg, err := ui.ResolvePackageByName(args[0])
	if err != nil {
		return err
	}
	if err := config.RemoveShellPackageDir(pkg.ID, pkg.Provider); err != nil {
		return fmt.Errorf("removing shell.d: %w", err)
	}
	fmt.Printf("Unset shell snippet for %s (provider: %s)\n", pkg.Name, pkg.Provider)
	return nil
}
