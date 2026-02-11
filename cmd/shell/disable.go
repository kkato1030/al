package shell

import (
	"github.com/spf13/cobra"
)

// NewDisableCmd creates the shell disable command
func NewDisableCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "disable <package-name>",
		Short: "Disable package shell snippet for al activate (file is kept)",
		Args:  cobra.ExactArgs(1),
		RunE:  runDisable,
	}
	return cmd
}

func runDisable(cmd *cobra.Command, args []string) error {
	return setEnabled(args[0], false)
}
