package link

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/kkato1030/al/internal/config"
	"github.com/spf13/cobra"
)

// NewEditCmd creates the link edit command
func NewEditCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit <name>",
		Short: "Edit link.d content in EDITOR",
		Long:  "Open the link.d content (file or dir root) in EDITOR (default: vim).",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runEdit(args[0])
		},
	}
	return cmd
}

func runEdit(name string) error {
	entry, entryDir, err := config.GetLinkByName(name)
	if err != nil {
		return err
	}
	if entry == nil {
		return fmt.Errorf("link not found: %s", name)
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
