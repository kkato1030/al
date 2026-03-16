package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	packagecmd "github.com/kkato1030/al/cmd/package"
	"github.com/kkato1030/al/internal/config"
	"github.com/kkato1030/al/internal/output"
	"github.com/kkato1030/al/internal/prompt"
	"github.com/spf13/cobra"
)

// NewReviewCmd creates the review command
func NewReviewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "review",
		Short: "Resolve overdue packages (remove / promote / move / postpone)",
		Long:  "List packages past their review deadline and choose remove (uninstall), promote (move to stable), move (move to another profile), or postpone (extend review by same period). Postpone shows a confirmation prompt.",
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
		fmt.Println("No packages overdue for review")
		return nil
	}

	fmt.Printf("Review overdue: %d package(s). Decide for each: remove / promote / move / postpone.\n\n", len(overdue))
	scanner := bufio.NewScanner(os.Stdin)

	for _, pkg := range overdue {
		resolved := false
		for !resolved {
			fmt.Printf("Package: %s (profile: %s, provider: %s)\n", pkg.Name, pkg.Profile, pkg.Provider)
			fmt.Print("  [r]emove  [p]romote  [m]ove  [s]postpone: ")
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
			case "m", "move":
				profilesCfg, err := config.LoadProfilesConfig()
				if err != nil {
					return fmt.Errorf("load profiles: %w", err)
				}
				var available []config.ProfileConfig
				for _, p := range profilesCfg.Profiles {
					if p.Name != pkg.Profile {
						available = append(available, p)
					}
				}
				if len(available) == 0 {
					fmt.Fprintln(os.Stderr, "  No other profiles available.")
					continue
				}
				fmt.Println("  Select destination profile:")
				for i, p := range available {
					line := fmt.Sprintf("    %d. %s", i+1, p.Name)
					if p.Description != "" {
						line += fmt.Sprintf(" - %s", p.Description)
					}
					fmt.Println(line)
				}
				fmt.Print("  Profile number: ")
				if !scanner.Scan() {
					return scanner.Err()
				}
				input := strings.TrimSpace(scanner.Text())
				idx, err := strconv.Atoi(input)
				if err != nil || idx < 1 || idx > len(available) {
					fmt.Fprintln(os.Stderr, "  Invalid selection, back to choice.")
					continue
				}
				toProfile := available[idx-1].Name
				if err := packagecmd.RunPackageMoveFromConfig(pkg, toProfile); err != nil {
					return fmt.Errorf("move %s: %w", pkg.Name, err)
				}
				resolved = true
			default:
				// postpone (default; "s", "postpone", or any other input)
				ok, err := prompt.Confirm(os.Stderr, "  Postpone anyway? [y/N]: ")
				if err != nil {
					return err
				}
				if !ok {
					fmt.Fprintln(os.Stderr, "  Back to choice.")
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
				output.Info("  Review extended to %s", reviewBy.Format("2006-01-02"))
				resolved = true
			}
		}
		fmt.Println()
	}

	return nil
}
