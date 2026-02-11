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

// NewAddCmd creates the shell add command
func NewAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <package-name>",
		Short: "Create shell snippet configuration and open in editor",
		Long:  "Create the package's shell.d snippet file and open it in EDITOR (default: vim). File extension is inferred from SHELL.",
		Args:  cobra.ExactArgs(1),
		RunE:  runAdd,
	}
	return cmd
}

func runAdd(cmd *cobra.Command, args []string) error {
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

	// Create the snippet file if it doesn't exist
	if _, err := os.Stat(snippetPath); os.IsNotExist(err) {
		if err := os.WriteFile(snippetPath, []byte("# Shell configuration for "+pkg.Name+"\n"), 0644); err != nil {
			return err
		}
		fmt.Printf("Created shell snippet for %s at %s\n", pkg.Name, snippetPath)
	} else {
		fmt.Printf("Shell snippet already exists for %s at %s\n", pkg.Name, snippetPath)
	}

	// Open in editor
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
