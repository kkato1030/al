package packagecmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/kkato1030/al/internal/brewfile"
	"github.com/kkato1030/al/internal/config"
	"github.com/kkato1030/al/internal/provider"
	"github.com/spf13/cobra"
)

// NewPackageImportCmd creates the package import command.
func NewPackageImportCmd() *cobra.Command {
	var profile string
	var stage string
	var install bool
	var dryRun bool
	var overwrite bool
	var verbose bool

	cmd := &cobra.Command{
		Use:   "import [Brewfile]",
		Short: "Import packages from a Brewfile",
		Long:  "Parse a Brewfile (tap, brew, cask, mas) and register packages to a profile. By default only registers; use --install to install missing packages.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var brewfilePath string
			if len(args) > 0 {
				brewfilePath = args[0]
			} else {
				var err error
				brewfilePath, err = brewfile.ResolveBrewfilePath("")
				if err != nil {
					return err
				}
			}

			if _, err := os.Stat(brewfilePath); err != nil {
				if os.IsNotExist(err) {
					return fmt.Errorf("Brewfile not found: %s", brewfilePath)
				}
				return fmt.Errorf("Brewfile: %w", err)
			}

			appConfig, err := config.LoadAppConfig()
			if err != nil {
				return fmt.Errorf("error loading app config: %w", err)
			}

			finalProfile, err := buildProfileName(profile, stage, appConfig.DefaultProfile, appConfig.DefaultStage)
			if err != nil {
				return fmt.Errorf("error building profile name: %w", err)
			}
			if finalProfile == "" {
				return fmt.Errorf("profile is required. Use --profile or set default profile with 'al config set --default-profile <profile>'")
			}

			profileConfig, err := findProfileWithFallback(finalProfile, stage)
			if err != nil {
				return fmt.Errorf("error loading profile: %w", err)
			}
			if profileConfig == nil {
				return fmt.Errorf("profile '%s' does not exist. Add it first with 'al profile add'", finalProfile)
			}
			finalProfile = profileConfig.Name

			result, err := brewfile.ParseFile(brewfilePath)
			if err != nil {
				return fmt.Errorf("parse Brewfile: %w", err)
			}

			needBrew := false
			needMas := false
			for _, e := range result.Entries {
				if e.Provider == "brew" {
					needBrew = true
				}
				if e.Provider == "mas" {
					needMas = true
				}
			}
			if needBrew {
				pc, _ := config.GetProvider("brew")
				if pc == nil {
					return fmt.Errorf("provider 'brew' is required for this Brewfile. Add it first with 'al provider add brew'")
				}
			}
			if needMas {
				pc, _ := config.GetProvider("mas")
				if pc == nil {
					return fmt.Errorf("provider 'mas' is required for this Brewfile. Add it first with 'al provider add mas'")
				}
			}

			if verbose && len(result.Skipped) > 0 {
				for _, s := range result.Skipped {
					fmt.Fprintf(os.Stderr, "Skipped line %d (%s): %s\n", s.LineNum, s.Reason, strings.TrimSpace(s.Line))
				}
			}

			if dryRun {
				fmt.Printf("Would import %d packages to profile '%s'\n", len(result.Entries), finalProfile)
				brewCount, masCount := 0, 0
				for _, e := range result.Entries {
					if e.Provider == "brew" {
						brewCount++
					} else {
						masCount++
					}
					fmt.Printf("  %s %s (%s)\n", e.Provider, e.ID, e.Name)
				}
				fmt.Printf("  brew: %d, mas: %d\n", brewCount, masCount)
				if len(result.Skipped) > 0 {
					fmt.Printf("Skipped %d lines (use --verbose to see details).\n", len(result.Skipped))
				}
				return nil
			}

			packagesConfig, err := config.LoadPackagesConfig()
			if err != nil {
				return fmt.Errorf("error loading packages config: %w", err)
			}

			existing := make(map[string]bool)
			for _, p := range packagesConfig.Packages {
				key := p.Provider + ":" + p.Profile + ":" + p.ID
				existing[key] = true
			}

			var brewProv provider.Provider
			var masProv provider.Provider
			if needBrew {
				brewProv = provider.NewBrewProvider()
			}
			if needMas {
				masProv = provider.NewMasProvider()
			}

			imported := 0
			skipped := 0
			brewImported := 0
			masImported := 0

			for _, e := range result.Entries {
				key := e.Provider + ":" + finalProfile + ":" + e.ID
				if existing[key] && !overwrite {
					skipped++
					continue
				}


				if install {
					if e.Provider == "brew" && brewProv != nil {
						if err := brewProv.InstallPackage(e.ID); err != nil {
							return fmt.Errorf("install %s: %w", e.ID, err)
						}
					}
					if e.Provider == "mas" && masProv != nil {
						if err := masProv.InstallPackage(e.ID); err != nil {
							return fmt.Errorf("install %s: %w", e.ID, err)
						}
					}
				}

				pkg := config.PackageConfig{
					ID:          e.ID,
					Name:        e.Name,
					Provider:    e.Provider,
					Profile:     finalProfile,
					InstalledAt: time.Now(),
				}
				if overwrite {
					if err := config.AddOrUpdatePackage(pkg); err != nil {
						return fmt.Errorf("add or update package %s: %w", e.ID, err)
					}
				} else {
					if err := config.AddPackage(pkg); err != nil {
						return fmt.Errorf("add package %s: %w", e.ID, err)
					}
				}
				imported++
				if e.Provider == "brew" {
					brewImported++
				} else {
					masImported++
				}
				existing[key] = true
			}

			fmt.Printf("Imported %d packages (brew: %d, mas: %d)", imported, brewImported, masImported)
			if skipped > 0 {
				fmt.Printf(". Skipped %d (already registered)", skipped)
			}
			fmt.Println()
			if len(result.Skipped) > 0 {
				fmt.Printf("Skipped %d lines (vscode/go/cargo/...). Use --verbose to see details.\n", len(result.Skipped))
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&profile, "profile", "f", "", "Profile to register packages to (required)")
	cmd.Flags().StringVarP(&stage, "stage", "s", "", "Stage name (optional)")
	cmd.Flags().BoolVar(&install, "install", false, "Install packages that are not yet installed via brew/mas")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be imported without writing")
	cmd.Flags().BoolVar(&overwrite, "overwrite", false, "Overwrite existing entries with same id, provider, profile")
	cmd.Flags().BoolVar(&verbose, "verbose", false, "Show skipped lines (unsupported types)")

	return cmd
}
