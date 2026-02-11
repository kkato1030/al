package shell

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kkato1030/al/internal/config"
	"github.com/kkato1030/al/internal/ui"
	"github.com/spf13/cobra"
)

// NewShowCmd creates the shell show command
func NewShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <package-name>",
		Short: "Show package shell.d content and manifest",
		Args:  cobra.ExactArgs(1),
		RunE:  runShow,
	}
	return cmd
}

func runShow(cmd *cobra.Command, args []string) error {
	pkg, err := ui.ResolvePackageByName(args[0])
	if err != nil {
		return err
	}
	pkgDir, err := config.GetShellPackageDir(pkg.ID, pkg.Provider)
	if err != nil {
		return err
	}
	if _, err := os.Stat(pkgDir); os.IsNotExist(err) {
		fmt.Printf("No shell.d entry for package %s (provider: %s)\n", pkg.Name, pkg.Provider)
		fmt.Printf("Path: %s (directory does not exist)\n", pkgDir)
		return nil
	}
	manifest, err := config.LoadShellManifest(pkgDir)
	if err != nil {
		return fmt.Errorf("loading manifest: %w", err)
	}
	fmt.Printf("Package: %s (provider: %s)\n", pkg.Name, pkg.Provider)
	fmt.Printf("Path: %s\n", pkgDir)
	fmt.Printf("Enabled: %v\n", manifest.Enabled)
	if manifest.After != "" {
		fmt.Printf("After: %s\n", manifest.After)
	}
	ext := shellExtFromEnv()
	paths, err := config.SnippetFilesInDir(pkgDir, ext)
	if err != nil {
		return err
	}
	if len(paths) == 0 {
		fmt.Println("(no snippet files)")
		return nil
	}
	for _, p := range paths {
		fmt.Printf("--- %s\n", filepath.Base(p))
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	}
	return nil
}
