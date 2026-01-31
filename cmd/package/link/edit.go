package link

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/kkato1030/al/internal/config"
	"github.com/kkato1030/al/internal/ui"
	"github.com/spf13/cobra"
)

// NewEditCmd creates the package link edit command
func NewEditCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit <package_name>",
		Short: "Edit link.d content for a package in EDITOR",
		Long:  "Open the link.d content (file or dir) for the package in EDITOR (default: vim).",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runEdit(args[0])
		},
	}
	return cmd
}

func runEdit(packageName string) error {
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
