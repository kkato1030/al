package link

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/kkato1030/al/internal/config"
	"github.com/kkato1030/al/internal/ui"
	"github.com/spf13/cobra"
)

// NewEditCmd creates the link edit command
func NewEditCmd() *cobra.Command {
	var path string
	var pkgName string
	cmd := &cobra.Command{
		Use:   "edit --path <path> [--package <pkg>]",
		Short: "Edit link.d content in EDITOR",
		Long:  "Open the link.d content (file or dir root) in EDITOR (default: vim).",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if path == "" {
				return fmt.Errorf("--path is required")
			}
			return runEdit(path, pkgName)
		},
	}
	cmd.Flags().StringVar(&path, "path", "", "Path (symlink location) to edit")
	cmd.Flags().StringVar(&pkgName, "package", "", "Filter by package name (optional)")
	return cmd
}

func runEdit(userPath, pkgName string) error {
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
	contentPath := config.GetLinkContentPath(entryDir)
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}
	editorCmd := exec.Command(editor, contentPath)
	editorCmd.Stdin = os.Stdin
	editorCmd.Stdout = os.Stdout
	editorCmd.Stderr = os.Stderr
	if err := editorCmd.Run(); err != nil {
		return fmt.Errorf("running %s: %w", editor, err)
	}
	return nil
}
