package bootstrap

import (
	"fmt"

	"github.com/kkato1030/al/internal/config"
	"github.com/spf13/cobra"
)

// NewAddCmd creates the bootstrap add command
func NewAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Create the bootstrap script",
		Long:  "Create ~/.al/bootstrap/script.sh with default content if it does not exist. Does nothing if the script already exists.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAdd()
		},
	}
	return cmd
}

func runAdd() error {
	path, created, err := config.EnsureBootstrapScript()
	if err != nil {
		return err
	}
	if created {
		fmt.Printf("Created bootstrap script: %s\n", path)
	} else {
		fmt.Printf("Bootstrap script already exists: %s\n", path)
	}
	return nil
}
