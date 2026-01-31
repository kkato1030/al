package link

import (
	"fmt"

	"github.com/kkato1030/al/internal/config"
	"github.com/spf13/cobra"
)

// NewRemoveCmd creates the link remove command
func NewRemoveCmd() *cobra.Command {
	var purge bool
	cmd := &cobra.Command{
		Use:   "remove <name> [--purge]",
		Short: "Remove a link from link.d",
		Long:  "Remove the symlink and copy content back to the path (default). Use --purge to delete the link.d content without copy-back.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRemove(args[0], purge)
		},
	}
	cmd.Flags().BoolVar(&purge, "purge", false, "Delete link.d content without copy-back")
	return cmd
}

func runRemove(name string, purge bool) error {
	entry, entryDir, err := config.GetLinkByName(name)
	if err != nil {
		return err
	}
	if entry == nil {
		return fmt.Errorf("link not found: %s", name)
	}
	if err := config.RemoveLink(entry, entryDir, purge); err != nil {
		return err
	}
	verb := "Removed"
	if purge {
		verb = "Purged"
	}
	fmt.Printf("%s link %s\n", verb, entry.Name)
	return nil
}
