package bootstrap

import (
	"fmt"

	"github.com/kkato1030/al/internal/config"
	"github.com/spf13/cobra"
)

// NewRemoveCmd creates the bootstrap remove command
func NewRemoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Delete the bootstrap script",
		Long:  "Delete ~/.al/bootstrap/script.sh. The bootstrap directory is left in place.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRemove()
		},
	}
	return cmd
}

func runRemove() error {
	if err := config.RemoveBootstrapScript(); err != nil {
		return err
	}
	fmt.Println("Removed bootstrap script")
	return nil
}
