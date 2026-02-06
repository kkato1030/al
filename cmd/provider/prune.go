package provider

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/kkato1030/al/internal/config"
	"github.com/spf13/cobra"
)

// Taps that should not be removed by prune (core/cask are commonly required).
var protectedBrewTaps = map[string]bool{
	"homebrew/core": true,
	"homebrew/cask": true,
}

// NewProviderPruneCmd creates the provider prune command.
// For brew: removes taps that are registered in al but have no package in packages.json
// that references them (no formula/cask with owner/repo/toolname form). Issue #50.
func NewProviderPruneCmd() *cobra.Command {
	var providerName string
	var dryRun bool
	var yes bool

	cmd := &cobra.Command{
		Use:   "prune",
		Short: "Prune provider resources (e.g. remove unused brew taps)",
		Long:  "For brew: remove taps that are in brew-taps.json but no package in packages has the form owner/repo/toolname (tap is unused). Protected taps (homebrew/core, homebrew/cask) are never removed. Use --dry-run to list what would be removed.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if providerName == "" {
				providerName = "brew"
			}
			if providerName != "brew" {
				return fmt.Errorf("prune is only supported for provider 'brew' (got %q)", providerName)
			}
			return runProviderPruneBrew(dryRun, yes)
		},
	}

	cmd.Flags().StringVar(&providerName, "provider", "brew", "Provider to prune (only brew is supported)")
	cmd.Flags().StringVar(&providerName, "prv", "", "Short form of --provider")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "List taps that would be removed without removing them")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Do not ask for confirmation before removing taps")

	return cmd
}

// tapFromPackageID extracts tap name "owner/repo" from brew package ID if the name part is owner/repo/toolname.
func tapFromPackageID(id string) string {
	// ID is "formula:owner/repo/toolname" or "cask:owner/repo/toolname"
	idx := strings.Index(id, ":")
	if idx < 0 {
		return ""
	}
	name := id[idx+1:]
	// name "owner/repo/toolname" → tap "owner/repo"
	parts := strings.SplitN(name, "/", 3)
	if len(parts) >= 2 {
		return parts[0] + "/" + parts[1]
	}
	return ""
}

func runProviderPruneBrew(dryRun, yes bool) error {
	managed, err := config.LoadBrewTapsConfig()
	if err != nil {
		return fmt.Errorf("load brew taps config: %w", err)
	}

	// Build set of taps that are used by any brew package (packages have owner/repo/toolname form).
	packagesConfig, err := config.LoadPackagesConfig()
	if err != nil {
		return fmt.Errorf("load packages config: %w", err)
	}
	usedTaps := make(map[string]bool)
	for _, pkg := range packagesConfig.Packages {
		if pkg.Provider != "brew" {
			continue
		}
		if tap := tapFromPackageID(pkg.ID); tap != "" {
			usedTaps[tap] = true
		}
	}

	// Taps to remove: in managed list, not protected, not used by any package.
	var toRemove []string
	for _, t := range managed.Taps {
		if protectedBrewTaps[t] {
			continue
		}
		if usedTaps[t] {
			continue
		}
		toRemove = append(toRemove, t)
	}

	if len(toRemove) == 0 {
		fmt.Println("No unused brew taps to remove.")
		return nil
	}

	if dryRun {
		fmt.Println("Taps that would be removed (use 'al provider prune' without --dry-run to remove):")
		for _, t := range toRemove {
			fmt.Printf("  %s\n", t)
		}
		return nil
	}

	if !yes {
		fmt.Printf("The following %d tap(s) will be removed (untap):\n", len(toRemove))
		for _, t := range toRemove {
			fmt.Printf("  %s\n", t)
		}
		fmt.Print("Continue? [y/N]: ")
		var answer string
		if _, err := fmt.Scanln(&answer); err != nil || (answer != "y" && answer != "Y") {
			fmt.Println("Aborted.")
			return nil
		}
	}

	for _, t := range toRemove {
		untap := exec.Command("brew", "untap", t)
		untap.Stdin = os.Stdin
		untap.Stdout = os.Stdout
		untap.Stderr = os.Stderr
		if err := untap.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to untap %s: %v\n", t, err)
		} else {
			fmt.Printf("Untapped %s\n", t)
		}
		// Remove from al's brew-taps.json so we no longer manage this tap
		if err := config.RemoveBrewTap(t); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to remove %s from brew-taps config: %v\n", t, err)
		}
	}
	return nil
}
