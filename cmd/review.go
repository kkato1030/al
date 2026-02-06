package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	packagecmd "github.com/kkato1030/al/cmd/package"
	"github.com/kkato1030/al/internal/config"
	"github.com/spf13/cobra"
)

// NewReviewCmd creates the review command
func NewReviewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "review",
		Short: "Resolve overdue packages (remove / promote / postpone)",
		Long:  "List packages past their review deadline and choose remove (uninstall), promote (move to stable), or postpone (extend review by same period). Postpone shows a confirmation prompt.",
		Args:  cobra.NoArgs,
		RunE:  runReview,
	}
	return cmd
}

func runReview(cmd *cobra.Command, args []string) error {
	overdue, err := config.GetOverduePackages()
	if err != nil {
		return fmt.Errorf("get overdue packages: %w", err)
	}
	if len(overdue) == 0 {
		fmt.Println("No packages overdue for review.")
		return nil
	}

	fmt.Printf("Review overdue: %d package(s). Decide for each: remove / promote / postpone.\n\n", len(overdue))
	scanner := bufio.NewScanner(os.Stdin)

	for _, pkg := range overdue {
		resolved := false
		for !resolved {
			fmt.Printf("Package: %s (profile: %s, provider: %s)\n", pkg.Name, pkg.Profile, pkg.Provider)
			fmt.Print("  [r]emove  [p]romote  [s]postpone: ")
			if !scanner.Scan() {
				return scanner.Err()
			}
			choice := strings.TrimSpace(strings.ToLower(scanner.Text()))
			if choice == "" {
				continue
			}

			switch choice {
			case "r", "remove":
				if err := packagecmd.RunPackageRemoveFromConfig(pkg, false, false); err != nil {
					return fmt.Errorf("remove %s: %w", pkg.Name, err)
				}
				resolved = true
			case "p", "promote":
				prof, err := config.GetProfile(pkg.Profile)
				if err != nil || prof == nil {
					return fmt.Errorf("get profile %s: %w", pkg.Profile, err)
				}
				if prof.PromoteTo == "" {
					return fmt.Errorf("profile %s has no promote_to (cannot promote)", pkg.Profile)
				}
				if err := packagecmd.RunPackageMoveFromConfig(pkg, prof.PromoteTo); err != nil {
					return fmt.Errorf("promote %s: %w", pkg.Name, err)
				}
				resolved = true
			default:
				// postpone (default; "s", "postpone", or any other input)
				fmt.Print("  Really? Do you really need it? [y/N]: ")
				if !scanner.Scan() {
					return scanner.Err()
				}
				confirm := strings.TrimSpace(strings.ToLower(scanner.Text()))
				if confirm != "y" && confirm != "yes" {
					fmt.Println("  Back to choice.")
					continue
				}
				days, hasReview, err := config.GetReviewDays(pkg.Profile)
				if err != nil || !hasReview || days <= 0 {
					return fmt.Errorf("profile %s has no review_days", pkg.Profile)
				}
				reviewBy := time.Now().AddDate(0, 0, days)
				if err := config.SetPackageReviewBy(pkg.ID, pkg.Provider, pkg.Profile, reviewBy); err != nil {
					return fmt.Errorf("postpone %s: %w", pkg.Name, err)
				}
				fmt.Printf("  Review extended to %s.\n", reviewBy.Format("2006-01-02"))
				resolved = true
			}
		}
		fmt.Println()
	}

	return nil
}
