package shell

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kkato1030/al/internal/config"
	"github.com/kkato1030/al/internal/ui"
	"github.com/spf13/cobra"
)

// NewSetCmd creates the shell set command
func NewSetCmd() *cobra.Command {
	var after string
	cmd := &cobra.Command{
		Use:   "set <package-name> <command_string> [--after <dep-package>]",
		Short: "Set package shell snippet content and load order",
		Long:  "Set shell.d snippet content from <command_string> and optionally --after <dep-package> for load order.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSet(args[0], args[1], after)
		},
	}
	cmd.Flags().StringVar(&after, "after", "", "Package (by name) that this snippet should load after")
	return cmd
}

func runSet(packageName, commandString, afterPackageName string) error {
	pkg, err := ui.ResolvePackageByName(packageName)
	if err != nil {
		return err
	}
	if err := config.EnsureShellPackageDir(pkg.ID, pkg.Provider); err != nil {
		return err
	}
	pkgDir, err := config.GetShellPackageDir(pkg.ID, pkg.Provider)
	if err != nil {
		return err
	}
	ext := shellExtFromEnv()
	snippetPath := filepath.Join(pkgDir, "snippet"+ext)
	content := []byte(commandString)
	if len(content) > 0 && content[len(content)-1] != '\n' {
		content = append(content, '\n')
	}
	if err := os.WriteFile(snippetPath, content, 0644); err != nil {
		return err
	}
	manifest, err := config.LoadShellManifest(pkgDir)
	if err != nil {
		return err
	}
	if afterPackageName != "" {
		depPkg, err := ui.ResolvePackageByName(afterPackageName)
		if err != nil {
			return fmt.Errorf("resolving --after package: %w", err)
		}
		manifest.After = config.PackageDirName(depPkg.ID, depPkg.Provider)
	}
	if err := config.SaveShellManifest(pkgDir, manifest); err != nil {
		return err
	}
	fmt.Printf("Set shell snippet for %s (provider: %s, profile: %s)\n", pkg.Name, pkg.Provider, pkg.Profile)
	return nil
}
