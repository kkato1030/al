package bootstrap

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/kkato1030/al/internal/config"
	"github.com/spf13/cobra"
)

// NewEditCmd creates the bootstrap edit command
func NewEditCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit",
		Short: "Edit the bootstrap script in EDITOR",
		Long:  "Open ~/.al/bootstrap/script.sh in EDITOR (default: vim). Creates the script with default content if it does not exist.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runEdit()
		},
	}
	return cmd
}

func runEdit() error {
	_, _, err := config.EnsureBootstrapScript()
	if err != nil {
		return err
	}
	path, err := config.GetBootstrapScriptPath()
	if err != nil {
		return err
	}
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}
	editorCmd := exec.Command(editor, path)
	editorCmd.Stdin = os.Stdin
	editorCmd.Stdout = os.Stdout
	editorCmd.Stderr = os.Stderr
	if err := editorCmd.Run(); err != nil {
		return fmt.Errorf("running %s: %w", editor, err)
	}
	return nil
}
