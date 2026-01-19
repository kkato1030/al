package cmd

import (
	"github.com/kkato1030/al/cmd/package"
	"github.com/kkato1030/al/cmd/profile"
	"github.com/kkato1030/al/cmd/provider"
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
	rootCmd.AddCommand(provider.NewProviderCmd())
	rootCmd.AddCommand(profile.NewProfileCmd())
	rootCmd.AddCommand(packagecmd.NewPackageCmd())

	return rootCmd
}
