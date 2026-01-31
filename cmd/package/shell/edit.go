package shell

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/kkato1030/al/internal/config"
	"github.com/kkato1030/al/internal/ui"
	"github.com/spf13/cobra"
)

// NewEditCmd creates the shell edit command
func NewEditCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit <package-name>",
		Short: "Edit package shell snippet in EDITOR",
		Long:  "Open the package's shell.d snippet file in EDITOR (default: vim). File extension is inferred from SHELL.",
		Args:  cobra.ExactArgs(1),
		RunE:  runEdit,
	}
	return cmd
}

func runEdit(cmd *cobra.Command, args []string) error {
	pkg, err := ui.ResolvePackageByName(args[0])
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
	if _, err := os.Stat(snippetPath); os.IsNotExist(err) {
		if err := os.WriteFile(snippetPath, nil, 0644); err != nil {
			return err
		}
	}
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}
	editorCmd := exec.Command(editor, snippetPath)
	editorCmd.Stdin = os.Stdin
	editorCmd.Stdout = os.Stdout
	editorCmd.Stderr = os.Stderr
	if err := editorCmd.Run(); err != nil {
		return fmt.Errorf("running %s: %w", editor, err)
	}
	return nil
}
