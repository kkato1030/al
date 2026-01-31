package shell

import (
	"github.com/spf13/cobra"
)

// NewEnableCmd creates the shell enable command
func NewEnableCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "enable <package-name>",
		Short: "Enable package shell snippet for al activate",
		Args:  cobra.ExactArgs(1),
		RunE:  runEnable,
	}
	return cmd
}

func runEnable(cmd *cobra.Command, args []string) error {
	return setEnabled(args[0], true)
}
