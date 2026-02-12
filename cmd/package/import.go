package packagecmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/kkato1030/al/internal/brewfile"
	"github.com/kkato1030/al/internal/config"
	"github.com/kkato1030/al/internal/provider"
	"github.com/spf13/cobra"
)

// detectInstalledPackages auto-detects installed packages from brew and mas
func detectInstalledPackages(interactive bool, profileName string) (*brewfile.ParseResult, error) {
	result := &brewfile.ParseResult{
		Entries: []brewfile.Entry{},
		Skipped: []brewfile.SkippedLine{},
	}

	// Cache mas list output for efficiency (used by getMasAppName)
	masListCache, err := cacheMasList()
	if err != nil {
		masListCache = make(map[string]string) // empty cache on error
	}

	// Check if brew provider is registered
	brewProvConfig, _ := config.GetProvider("brew")
	if brewProvConfig != nil {
		brewProv := provider.NewBrewProvider()
		installed, err := brewProv.CheckInstalled()
		if err == nil && installed {
			// Get list of installed packages
			installedPkgs, err := brewProv.ListInstalled()
			if err != nil {
				return nil, fmt.Errorf("failed to list brew packages: %w", err)
			}

			// Convert to entries
			// The map values are just boolean markers (true), indicating presence
			for pkgID := range installedPkgs {
				// Determine name from ID
				name := pkgID
				if strings.HasPrefix(pkgID, "formula:") {
					name = strings.TrimPrefix(pkgID, "formula:")
				} else if strings.HasPrefix(pkgID, "cask:") {
					name = strings.TrimPrefix(pkgID, "cask:")
				} else if strings.HasPrefix(pkgID, "tap:") {
					name = strings.TrimPrefix(pkgID, "tap:")
				}

				// Skip entries without proper format (backward compatibility entries)
				if !strings.Contains(pkgID, ":") {
					continue
				}

				entry := brewfile.Entry{
					ID:       pkgID,
					Name:     name,
					Provider: "brew",
				}

				// Interactive mode: ask user for each package
				if interactive {
					if !promptForPackage(entry, profileName) {
						continue
					}
				}

				result.Entries = append(result.Entries, entry)
			}
		}
	}

	// Check if mas provider is registered
	masProvConfig, _ := config.GetProvider("mas")
	if masProvConfig != nil {
		masProv := provider.NewMasProvider()
		installed, err := masProv.CheckInstalled()
		if err == nil && installed {
			// Get list of installed apps
			installedApps, err := masProv.ListInstalled()
			if err != nil {
				return nil, fmt.Errorf("failed to list mas apps: %w", err)
			}

			// Convert to entries
			// The map values are just boolean markers (true), indicating presence
			for appID := range installedApps {
				// Get the app name from cached mas list output
				name := masListCache[appID]
				if name == "" {
					name = appID // fallback to ID if name not in cache
				}

				entry := brewfile.Entry{
					ID:       appID,
					Name:     name,
					Provider: "mas",
				}

				// Interactive mode: ask user for each package
				if interactive {
					if !promptForPackage(entry, profileName) {
						continue
					}
				}

				result.Entries = append(result.Entries, entry)
			}
		}
	}

	return result, nil
}

// cacheMasList runs `mas list` once and returns a map of appID -> appName
func cacheMasList() (map[string]string, error) {
	cache := make(map[string]string)

	masProv := provider.NewMasProvider()

	// Check if mas is installed
	installed, err := masProv.CheckInstalled()
	if err != nil || !installed {
		return cache, fmt.Errorf("mas not installed")
	}

	// Run mas list to get all apps
	cmd := exec.Command("mas", "list")
	output, err := cmd.Output()
	if err != nil {
		return cache, fmt.Errorf("failed to run mas list: %w", err)
	}

	// Parse output to build cache
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Format: "123456789 App Name (version)"
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			appID := fields[0]
			// Extract name (everything between ID and potential version)
			nameParts := fields[1:]
			name := strings.Join(nameParts, " ")
			// Remove version info if present (e.g., "(1.2.3)")
			if idx := strings.Index(name, "("); idx >= 0 {
				name = strings.TrimSpace(name[:idx])
			}
			cache[appID] = name
		}
	}

	return cache, nil
}

// promptForPackage asks the user whether to import a package
func promptForPackage(entry brewfile.Entry, profileName string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Import %s %s (%s) to profile '%s'? [y/N]: ", entry.Provider, entry.Name, entry.ID, profileName)

	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}

// NewPackageImportCmd creates the package import command.
func NewPackageImportCmd() *cobra.Command {
	var profile string
	var stage string
	var install bool
	var dryRun bool
	var overwrite bool
	var verbose bool
	var interactive bool

	cmd := &cobra.Command{
		Use:   "import [Brewfile]",
		Short: "Import packages from a Brewfile or auto-detect from brew/mas",
		Long:  "Parse a Brewfile (tap, brew, cask, mas) and register packages to a profile. If no Brewfile is specified and brew/mas are installed, auto-detect installed packages. By default only registers; use --install to install missing packages.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var brewfilePath string
			var autoDetect bool

			if len(args) > 0 {
				brewfilePath = args[0]
				autoDetect = false
			} else {
				// Try to resolve Brewfile path
				var err error
				brewfilePath, err = brewfile.ResolveBrewfilePath("")
				if err != nil || brewfilePath == "" {
					// No Brewfile found, enable auto-detection
					autoDetect = true
				} else {
					// Check if the resolved Brewfile exists
					if _, err := os.Stat(brewfilePath); err != nil {
						if os.IsNotExist(err) {
							// Brewfile doesn't exist, enable auto-detection
							autoDetect = true
						} else {
							return fmt.Errorf("Brewfile: %w", err)
						}
					}
				}
			}

			// If not auto-detecting, validate Brewfile exists
			if !autoDetect {
				if _, err := os.Stat(brewfilePath); err != nil {
					if os.IsNotExist(err) {
						return fmt.Errorf("Brewfile not found: %s", brewfilePath)
					}
					return fmt.Errorf("Brewfile: %w", err)
				}
			}

			appConfig, err := config.LoadAppConfig()
			if err != nil {
				return fmt.Errorf("error loading app config: %w", err)
			}

			finalProfile, appliedStage, err := buildProfileName(profile, stage, appConfig.DefaultProfile, appConfig.DefaultStage)
			if err != nil {
				return fmt.Errorf("error building profile name: %w", err)
			}
			if finalProfile == "" {
				return fmt.Errorf("profile is required. Use --profile or set default profile with 'al config set --default-profile <profile>'")
			}

			profileConfig, err := findProfileWithFallback(finalProfile, appliedStage)
			if err != nil {
				return fmt.Errorf("error loading profile: %w", err)
			}
			if profileConfig == nil {
				return fmt.Errorf("profile '%s' does not exist. Add it first with 'al profile add'", finalProfile)
			}
			finalProfile = profileConfig.Name

			var result *brewfile.ParseResult
			if autoDetect {
				// Auto-detect installed packages from brew and mas
				result, err = detectInstalledPackages(interactive, finalProfile)
				if err != nil {
					return fmt.Errorf("auto-detect packages: %w", err)
				}
				if len(result.Entries) == 0 {
					fmt.Println("No packages detected from brew or mas. Make sure brew/mas are installed and registered.")
					return nil
				}
			} else {
				result, err = brewfile.ParseFile(brewfilePath)
				if err != nil {
					return fmt.Errorf("parse Brewfile: %w", err)
				}
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
				sourceInfo := ""
				if autoDetect {
					sourceInfo = " (auto-detected from brew/mas)"
				} else {
					sourceInfo = fmt.Sprintf(" from %s", brewfilePath)
				}
				fmt.Printf("Would import %d packages to profile '%s'%s\n", len(result.Entries), finalProfile, sourceInfo)
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
			tapsImported := 0

			for _, e := range result.Entries {
				// Brew taps are managed by provider brew, not as packages (issue #50).
				if e.Provider == "brew" && strings.HasPrefix(e.ID, "tap:") {
					tapName := strings.TrimPrefix(e.ID, "tap:")
					hasTap, err := config.HasBrewTap(tapName)
					if err != nil {
						return fmt.Errorf("load brew taps: %w", err)
					}
					if !hasTap {
						if install && brewProv != nil {
							if err := brewProv.InstallPackage(e.ID); err != nil {
								return fmt.Errorf("install tap %s: %w", tapName, err)
							}
						}
						if err := config.AddBrewTap(tapName); err != nil {
							return fmt.Errorf("add brew tap %s: %w", tapName, err)
						}
						tapsImported++
					}
					continue
				}

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
				if days, hasReview, _ := config.GetReviewDays(finalProfile); hasReview && days > 0 {
					reviewBy := time.Now().AddDate(0, 0, days)
					pkg.ReviewBy = &reviewBy
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

			sourceInfo := ""
			if autoDetect {
				sourceInfo = " (auto-detected from brew/mas)"
			} else {
				sourceInfo = fmt.Sprintf(" from %s", brewfilePath)
			}

			fmt.Printf("Imported %d packages (brew: %d, mas: %d)%s", imported, brewImported, masImported, sourceInfo)
			if tapsImported > 0 {
				fmt.Printf(", %d tap(s) (managed by provider brew)", tapsImported)
			}
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

	cmd.Flags().StringVar(&profile, "profile", "", "Profile to register packages to (required)")
	cmd.Flags().StringVar(&profile, "prf", "", "Short form of --profile")
	cmd.Flags().StringVarP(&stage, "stage", "s", "", "Stage name (optional)")
	cmd.Flags().BoolVar(&install, "install", false, "Install packages that are not yet installed via brew/mas")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be imported without writing")
	cmd.Flags().BoolVar(&overwrite, "overwrite", false, "Overwrite existing entries with same id, provider, profile")
	cmd.Flags().BoolVar(&verbose, "verbose", false, "Show skipped lines (unsupported types)")
	cmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Interactive mode: choose which packages to import")

	return cmd
}
