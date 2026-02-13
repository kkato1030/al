package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/kkato1030/al/internal/config"
	"github.com/kkato1030/al/internal/provider"
	"github.com/spf13/cobra"
)

// CheckStatus represents the status of a diagnostic check
type CheckStatus string

const (
	StatusOK    CheckStatus = "OK"
	StatusWarn  CheckStatus = "WARN"
	StatusError CheckStatus = "ERROR"
)

// CheckResult represents the result of a diagnostic check
type CheckResult struct {
	Status  CheckStatus
	Message string
}

// NewDoctorCmd creates the doctor command
func NewDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Check for broken or inconsistent state",
		Long:  "Diagnose environment issues without modifying system state. Checks provider availability, broken links, shell.d consistency, and more.",
		Args:  cobra.NoArgs,
		RunE:  runDoctor,
	}
}

func runDoctor(cmd *cobra.Command, args []string) error {
	results := []CheckResult{}

	// Check configuration directory
	results = append(results, checkConfigDir()...)

	// Check configuration files
	results = append(results, checkConfigFiles()...)

	// Check providers
	results = append(results, checkProviders()...)

	// Check links
	results = append(results, checkLinks()...)

	// Check shell.d
	results = append(results, checkShellD()...)

	// Check for overdue packages
	results = append(results, checkOverduePackages()...)

	// Check for invalid profile references
	results = append(results, checkInvalidProfileReferences()...)

	// Print results
	printResults(results)

	// Return error if any ERROR status found
	for _, r := range results {
		if r.Status == StatusError {
			return fmt.Errorf("doctor found errors")
		}
	}

	return nil
}

func checkConfigDir() []CheckResult {
	results := []CheckResult{}
	configDir, err := config.GetConfigDir()
	if err != nil {
		results = append(results, CheckResult{
			Status:  StatusError,
			Message: fmt.Sprintf("config directory: cannot determine path: %v", err),
		})
		return results
	}

	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		results = append(results, CheckResult{
			Status:  StatusError,
			Message: fmt.Sprintf("config directory %s does not exist (run 'al init')", configDir),
		})
		return results
	}

	results = append(results, CheckResult{
		Status:  StatusOK,
		Message: fmt.Sprintf("config directory %s exists", configDir),
	})

	// Check required subdirectories
	subdirs := []string{"link.d", "shell.d", "bootstrap"}
	for _, subdir := range subdirs {
		path := filepath.Join(configDir, subdir)
		if _, err := os.Stat(path); err == nil {
			results = append(results, CheckResult{
				Status:  StatusOK,
				Message: fmt.Sprintf("directory %s/ exists", subdir),
			})
		}
		// Not an error if subdirectories don't exist - they're created on demand
	}

	return results
}

func checkConfigFiles() []CheckResult {
	results := []CheckResult{}

	// Check config.json
	if cfg, err := config.LoadAppConfig(); err != nil {
		results = append(results, CheckResult{
			Status:  StatusWarn,
			Message: fmt.Sprintf("config.json: cannot load: %v", err),
		})
	} else {
		results = append(results, CheckResult{
			Status:  StatusOK,
			Message: "config.json is valid",
		})
		// Check for valid default profile if set
		if cfg.DefaultProfile != "" {
			if prof, err := config.GetProfile(cfg.DefaultProfile); err != nil || prof == nil {
				results = append(results, CheckResult{
					Status:  StatusWarn,
					Message: fmt.Sprintf("default profile '%s' not found", cfg.DefaultProfile),
				})
			}
		}
	}

	// Check providers.json
	if providersCfg, err := config.LoadProvidersConfig(); err != nil {
		results = append(results, CheckResult{
			Status:  StatusWarn,
			Message: fmt.Sprintf("providers.json: cannot load: %v", err),
		})
	} else {
		if len(providersCfg.Providers) == 0 {
			results = append(results, CheckResult{
				Status:  StatusWarn,
				Message: "no providers configured",
			})
		} else {
			results = append(results, CheckResult{
				Status:  StatusOK,
				Message: fmt.Sprintf("providers.json is valid (%d provider(s))", len(providersCfg.Providers)),
			})
		}
	}

	// Check profiles.json
	if profilesCfg, err := config.LoadProfilesConfig(); err != nil {
		results = append(results, CheckResult{
			Status:  StatusWarn,
			Message: fmt.Sprintf("profiles.json: cannot load: %v", err),
		})
	} else {
		if len(profilesCfg.Profiles) == 0 {
			results = append(results, CheckResult{
				Status:  StatusWarn,
				Message: "no profiles configured",
			})
		} else {
			results = append(results, CheckResult{
				Status:  StatusOK,
				Message: fmt.Sprintf("profiles.json is valid (%d profile(s))", len(profilesCfg.Profiles)),
			})
		}
	}

	// Check packages.json
	if packagesCfg, err := config.LoadPackagesConfig(); err != nil {
		results = append(results, CheckResult{
			Status:  StatusWarn,
			Message: fmt.Sprintf("packages.json: cannot load: %v", err),
		})
	} else {
		results = append(results, CheckResult{
			Status:  StatusOK,
			Message: fmt.Sprintf("packages.json is valid (%d package(s))", len(packagesCfg.Packages)),
		})
	}

	return results
}

func checkProviders() []CheckResult {
	results := []CheckResult{}

	// Check brew
	brewProvider := provider.NewBrewProvider()
	if installed, err := brewProvider.CheckInstalled(); err != nil {
		results = append(results, CheckResult{
			Status:  StatusWarn,
			Message: fmt.Sprintf("brew: check failed: %v", err),
		})
	} else if installed {
		results = append(results, CheckResult{
			Status:  StatusOK,
			Message: "brew is available",
		})
	} else {
		results = append(results, CheckResult{
			Status:  StatusWarn,
			Message: "brew is not installed",
		})
	}

	// Check mas
	masProvider := provider.NewMasProvider()
	if installed, err := masProvider.CheckInstalled(); err != nil {
		results = append(results, CheckResult{
			Status:  StatusWarn,
			Message: fmt.Sprintf("mas: check failed: %v", err),
		})
	} else if installed {
		results = append(results, CheckResult{
			Status:  StatusOK,
			Message: "mas is available",
		})
	} else {
		results = append(results, CheckResult{
			Status:  StatusWarn,
			Message: "mas is not installed",
		})
	}

	// Check gh (GitHub CLI) - optional but useful
	ghCmd := exec.Command("gh", "--version")
	if err := ghCmd.Run(); err == nil {
		results = append(results, CheckResult{
			Status:  StatusOK,
			Message: "gh (GitHub CLI) is available",
		})
	} else {
		results = append(results, CheckResult{
			Status:  StatusWarn,
			Message: "gh (GitHub CLI) is not installed (optional for backup/sync)",
		})
	}

	return results
}

func checkLinks() []CheckResult {
	results := []CheckResult{}

	links, err := config.ListLinks("", "")
	if err != nil {
		results = append(results, CheckResult{
			Status:  StatusWarn,
			Message: fmt.Sprintf("link.d: cannot list: %v", err),
		})
		return results
	}

	if len(links) == 0 {
		// Not an error - just no links configured
		return results
	}

	brokenCount := 0
	for _, link := range links {
		userPath := link.Manifest.UserPath
		fi, err := os.Lstat(userPath)
		if err != nil {
			if os.IsNotExist(err) {
				results = append(results, CheckResult{
					Status:  StatusWarn,
					Message: fmt.Sprintf("link %s: symlink target does not exist at %s", link.Name, userPath),
				})
				brokenCount++
			} else {
				results = append(results, CheckResult{
					Status:  StatusWarn,
					Message: fmt.Sprintf("link %s: cannot stat %s: %v", link.Name, userPath, err),
				})
				brokenCount++
			}
			continue
		}

		// Check if it's actually a symlink
		if fi.Mode()&os.ModeSymlink == 0 {
			results = append(results, CheckResult{
				Status:  StatusWarn,
				Message: fmt.Sprintf("link %s: %s is not a symlink", link.Name, userPath),
			})
			brokenCount++
			continue
		}

		// Check if symlink target exists
		if _, err := os.Stat(userPath); err != nil {
			results = append(results, CheckResult{
				Status:  StatusWarn,
				Message: fmt.Sprintf("link %s: symlink %s is broken", link.Name, userPath),
			})
			brokenCount++
		}
	}

	if brokenCount == 0 {
		results = append(results, CheckResult{
			Status:  StatusOK,
			Message: fmt.Sprintf("all %d link(s) are valid", len(links)),
		})
	}

	return results
}

func checkShellD() []CheckResult {
	results := []CheckResult{}

	dirNames, err := config.ListShellPackageDirNames()
	if err != nil {
		results = append(results, CheckResult{
			Status:  StatusWarn,
			Message: fmt.Sprintf("shell.d: cannot list: %v", err),
		})
		return results
	}

	if len(dirNames) == 0 {
		// Not an error - just no shell entries
		return results
	}

	// Check for cycles in dependency graph (using GetEnabledShellEntriesInOrder)
	// Try common shell extensions
	for _, ext := range []string{".zsh", ".bash"} {
		_, err := config.GetEnabledShellEntriesInOrder(ext)
		if err != nil {
			if strings.Contains(err.Error(), "cycle") {
				results = append(results, CheckResult{
					Status:  StatusError,
					Message: fmt.Sprintf("shell.d: dependency cycle detected: %v", err),
				})
				return results
			}
		}
	}

	// Check for missing files in enabled entries
	shellDir, _ := config.GetShellDir()
	missingCount := 0
	for _, dirName := range dirNames {
		pkgDir := filepath.Join(shellDir, dirName)
		manifest, err := config.LoadShellManifest(pkgDir)
		if err != nil {
			results = append(results, CheckResult{
				Status:  StatusWarn,
				Message: fmt.Sprintf("shell.d/%s: cannot load manifest: %v", dirName, err),
			})
			missingCount++
			continue
		}

		// If enabled, check for at least one snippet file
		if manifest.Enabled {
			entries, err := os.ReadDir(pkgDir)
			if err != nil {
				results = append(results, CheckResult{
					Status:  StatusWarn,
					Message: fmt.Sprintf("shell.d/%s: cannot read directory: %v", dirName, err),
				})
				missingCount++
				continue
			}

			hasSnippet := false
			for _, e := range entries {
				if !e.IsDir() && (strings.HasSuffix(e.Name(), ".zsh") || strings.HasSuffix(e.Name(), ".bash")) {
					hasSnippet = true
					break
				}
			}

			if !hasSnippet {
				results = append(results, CheckResult{
					Status:  StatusWarn,
					Message: fmt.Sprintf("shell.d/%s: enabled but no snippet files found", dirName),
				})
				missingCount++
			}
		}
	}

	if missingCount == 0 {
		results = append(results, CheckResult{
			Status:  StatusOK,
			Message: fmt.Sprintf("shell.d has %d package(s), no issues found", len(dirNames)),
		})
	}

	return results
}

func checkOverduePackages() []CheckResult {
	results := []CheckResult{}

	overdue, err := config.GetOverduePackages()
	if err != nil {
		results = append(results, CheckResult{
			Status:  StatusWarn,
			Message: fmt.Sprintf("review: cannot check overdue packages: %v", err),
		})
		return results
	}

	if len(overdue) > 0 {
		// Build a list of overdue package names
		names := make([]string, 0, len(overdue))
		for _, pkg := range overdue {
			names = append(names, pkg.Name)
		}
		results = append(results, CheckResult{
			Status:  StatusWarn,
			Message: fmt.Sprintf("trial packages review expired (%d): %s (run 'al review')", len(overdue), strings.Join(names, ", ")),
		})
	}

	return results
}

func checkInvalidProfileReferences() []CheckResult {
	results := []CheckResult{}

	profilesCfg, err := config.LoadProfilesConfig()
	if err != nil {
		// Already reported in checkConfigFiles
		return results
	}

	packagesCfg, err := config.LoadPackagesConfig()
	if err != nil {
		// Already reported in checkConfigFiles
		return results
	}

	// Build set of valid profile names
	validProfiles := make(map[string]bool)
	for _, prof := range profilesCfg.Profiles {
		validProfiles[prof.Name] = true
	}

	// Check each package's profile reference
	invalidCount := 0
	for _, pkg := range packagesCfg.Packages {
		if !validProfiles[pkg.Profile] {
			results = append(results, CheckResult{
				Status:  StatusWarn,
				Message: fmt.Sprintf("package %s: references non-existent profile '%s'", pkg.Name, pkg.Profile),
			})
			invalidCount++
		}
	}

	// Check promote_to references
	for _, prof := range profilesCfg.Profiles {
		if prof.PromoteTo != "" && !validProfiles[prof.PromoteTo] {
			results = append(results, CheckResult{
				Status:  StatusWarn,
				Message: fmt.Sprintf("profile %s: promote_to references non-existent profile '%s'", prof.Name, prof.PromoteTo),
			})
			invalidCount++
		}
	}

	if invalidCount == 0 && len(packagesCfg.Packages) > 0 {
		results = append(results, CheckResult{
			Status:  StatusOK,
			Message: "all profile references are valid",
		})
	}

	return results
}

func printResults(results []CheckResult) {
	okCount := 0
	warnCount := 0
	errorCount := 0

	for _, r := range results {
		switch r.Status {
		case StatusOK:
			okCount++
		case StatusWarn:
			warnCount++
		case StatusError:
			errorCount++
		}
	}

	// Print header
	fmt.Println("Running diagnostics...")
	fmt.Println()

	// Print results
	for _, r := range results {
		prefix := ""
		switch r.Status {
		case StatusOK:
			prefix = "[OK]   "
		case StatusWarn:
			prefix = "[WARN] "
		case StatusError:
			prefix = "[ERROR]"
		}
		fmt.Printf("%s %s\n", prefix, r.Message)
	}

	// Print summary
	fmt.Println()
	fmt.Printf("Summary: %d OK, %d warnings, %d errors\n", okCount, warnCount, errorCount)
}
