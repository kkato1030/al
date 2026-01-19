package cmd

import (
	"github.com/spf13/cobra"
)

var version = "0.1.0"

// NewRootCmd creates the root command
func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "al",
		Short: "Mac Management Tools",
		Long:  "al - Mac Management Tools",
	}

	rootCmd.AddCommand(NewVersionCmd())
	rootCmd.AddCommand(NewProviderCmd())

	return rootCmd
}
