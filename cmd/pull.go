package cmd

import (
	"github.com/spf13/cobra"
)

// NewPullCmd creates the pull command
func NewPullCmd() *cobra.Command {
	var repo string
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "pull",
		Short: "Fetch and merge changes from the remote backup into ~/.al",
		Long:  "Fetch and merge changes from the remote backup repository into ~/.al (AL_HOME). If merge conflicts occur, they are reported and the user is guided to resolve them manually.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBackupPull(repo, dryRun)
		},
	}

	cmd.Flags().StringVar(&repo, "repo", "", "Override backup repository (owner/repo, e.g. kkato1030/dotal)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be pulled without actually fetching or merging")

	return cmd
}
