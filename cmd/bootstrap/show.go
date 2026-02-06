package bootstrap

import (
	"os"

	"github.com/kkato1030/al/internal/config"
	"github.com/spf13/cobra"
)

// NewShowCmd creates the bootstrap show command
func NewShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show the bootstrap script content",
		Long:  "Print the content of ~/.al/bootstrap/script.sh to stdout.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runShow()
		},
	}
	return cmd
}

func runShow() error {
	data, err := config.ReadBootstrapScript()
	if err != nil {
		return err
	}
	_, err = os.Stdout.Write(data)
	return err
}
