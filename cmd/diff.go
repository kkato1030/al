package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/kkato1030/al/internal/config"
	"github.com/kkato1030/al/internal/provider"
	"github.com/spf13/cobra"
)

// NewDiffCmd creates the diff command
func NewDiffCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "diff",
		Short: "Compare desired state and current state",
		Long:  "Shows the difference between packages defined in profiles (desired state) and packages currently installed on the system. Displays additions (+), removals (-), and upgrades (~) available.",
		RunE: func(cmd *cobra.Command, args []string) error {
			exitCode := runDiff()
			if exitCode != 0 {
				// Drift detected - we've already printed output
				// Return nil to avoid error message, but set exit code using os.Exit
				// This is acceptable here because we've finished all our work
				os.Exit(exitCode)
			}
			return nil
		},
	}

	return cmd
}

type diffResult struct {
	additions []packageDiff // packages in profiles but not installed
	removals  []packageDiff // packages installed but not in any profile
	upgrades  []packageDiff // packages installed with upgrades available
}

type packageDiff struct {
	provider string
	name     string
	id       string
}

func runDiff() int {
	// Load packages config to get desired state
	packagesConfig, err := config.LoadPackagesConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load packages config: %v\n", err)
		return 1
	}

	// Build a set of desired packages from profiles
	// Key: packageID, Value: package info
	desiredPackages := make(map[string]packageDiff)
	for _, pkg := range packagesConfig.Packages {
		// Exclude taps as they are managed by provider brew, not as packages
		if pkg.Provider == "brew" && strings.HasPrefix(pkg.ID, "tap:") {
			continue
		}
		// Exclude manual packages
		if pkg.Provider == "manual" {
			continue
		}
		desiredPackages[pkg.ID] = packageDiff{
			provider: pkg.Provider,
			name:     pkg.Name,
			id:       pkg.ID,
		}
	}

	// Get currently installed packages from system
	installedPackages, err := getInstalledPackages()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get installed packages: %v\n", err)
		return 1
	}

	// Get upgradable packages from system
	upgradablePackages, err := getUpgradablePackages()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get upgradable packages: %v\n", err)
		return 1
	}

	// Calculate diff
	result := calculateDiff(desiredPackages, installedPackages, upgradablePackages)

	// Display results
	displayDiff(result)

	// Return non-zero exit code if there are any differences
	if len(result.additions) > 0 || len(result.removals) > 0 || len(result.upgrades) > 0 {
		return 1
	}

	return 0
}

// getInstalledPackages returns all installed packages from brew and mas
func getInstalledPackages() (map[string]bool, error) {
	installed := make(map[string]bool)

	// Get brew packages
	brewProvider := provider.NewBrewProvider()
	brewInstalled, err := brewProvider.CheckInstalled()
	if err != nil {
		return nil, fmt.Errorf("failed to check brew installation: %w", err)
	}

	if brewInstalled {
		// Get installed formulas (using --installed-on-request to exclude dependencies)
		formulaCmd := exec.Command("brew", "list", "--formula", "--installed-on-request")
		formulaOutput, err := formulaCmd.Output()
		if err == nil {
			lines := strings.Split(strings.TrimSpace(string(formulaOutput)), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line != "" {
					installed["formula:"+line] = true
					installed[line] = true // for backward compatibility
				}
			}
		}

		// Get installed casks
		caskCmd := exec.Command("brew", "list", "--cask")
		caskOutput, err := caskCmd.Output()
		if err == nil {
			lines := strings.Split(strings.TrimSpace(string(caskOutput)), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line != "" {
					installed["cask:"+line] = true
				}
			}
		}
	}

	// Get mas packages
	masProvider := provider.NewMasProvider()
	masInstalled, err := masProvider.CheckInstalled()
	if err != nil {
		return nil, fmt.Errorf("failed to check mas installation: %w", err)
	}

	if masInstalled {
		masCmd := exec.Command("mas", "list")
		masOutput, err := masCmd.Output()
		if err == nil {
			lines := strings.Split(strings.TrimSpace(string(masOutput)), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}
				// Parse line format: "123456789 App Name"
				fields := strings.Fields(line)
				if len(fields) > 0 {
					appID := fields[0]
					installed[appID] = true
				}
			}
		}
	}

	return installed, nil
}

// getUpgradablePackages returns all upgradable packages from brew and mas
// Only returns packages that exist in the desired state (al list)
func getUpgradablePackages() (map[string]bool, error) {
	upgradable := make(map[string]bool)

	// Get brew upgradable packages
	brewProvider := provider.NewBrewProvider()
	brewInstalled, err := brewProvider.CheckInstalled()
	if err != nil {
		return nil, fmt.Errorf("failed to check brew installation: %w", err)
	}

	if brewInstalled {
		// Get outdated formulas
		formulaCmd := exec.Command("brew", "outdated", "--formula")
		formulaOutput, err := formulaCmd.Output()
		if err == nil {
			lines := strings.Split(strings.TrimSpace(string(formulaOutput)), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}
				// Parse format: "package (version) < version"
				fields := strings.Fields(line)
				if len(fields) > 0 {
					pkgName := fields[0]
					upgradable["formula:"+pkgName] = true
					upgradable[pkgName] = true
				}
			}
		}

		// Get outdated casks
		caskCmd := exec.Command("brew", "outdated", "--cask")
		caskOutput, err := caskCmd.Output()
		if err == nil {
			lines := strings.Split(strings.TrimSpace(string(caskOutput)), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}
				// Parse format: similar to formulas
				fields := strings.Fields(line)
				if len(fields) > 0 {
					pkgName := fields[0]
					upgradable["cask:"+pkgName] = true
				}
			}
		}
	}

	// Get mas upgradable packages
	masProvider := provider.NewMasProvider()
	masInstalled, err := masProvider.CheckInstalled()
	if err != nil {
		return nil, fmt.Errorf("failed to check mas installation: %w", err)
	}

	if masInstalled {
		masCmd := exec.Command("mas", "outdated")
		masOutput, err := masCmd.Output()
		if err == nil {
			lines := strings.Split(strings.TrimSpace(string(masOutput)), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}
				// Parse line format: "123456789 App Name (version -> newversion)"
				fields := strings.Fields(line)
				if len(fields) > 0 {
					appID := fields[0]
					upgradable[appID] = true
				}
			}
		}
	}

	return upgradable, nil
}

// calculateDiff compares desired and installed packages
func calculateDiff(desired map[string]packageDiff, installed map[string]bool, upgradable map[string]bool) diffResult {
	result := diffResult{
		additions: []packageDiff{},
		removals:  []packageDiff{},
		upgrades:  []packageDiff{},
	}

	// Find additions (in desired but not installed)
	for id, pkg := range desired {
		if !installed[id] {
			result.additions = append(result.additions, pkg)
		} else if upgradable[id] {
			// Package is installed and upgradable, and it's in desired state
			result.upgrades = append(result.upgrades, pkg)
		}
	}

	// Find removals (installed but not in desired)
	for id := range installed {
		// Check if this package is in desired state
		if _, ok := desired[id]; !ok {
			// Try to infer provider and name from ID
			var providerName, name string
			if strings.HasPrefix(id, "formula:") {
				providerName = "brew"
				name = strings.TrimPrefix(id, "formula:")
			} else if strings.HasPrefix(id, "cask:") {
				providerName = "brew"
				name = strings.TrimPrefix(id, "cask:")
			} else if strings.Contains(id, ":") {
				// Skip prefixed IDs that are already handled
				continue
			} else {
				// Check if it's a mas app (numeric ID)
				var appID int
				if _, err := fmt.Sscanf(id, "%d", &appID); err == nil {
					providerName = "mas"
					name = id
				} else {
					// Assume it's a brew formula
					providerName = "brew"
					name = id
				}
			}
			result.removals = append(result.removals, packageDiff{
				provider: providerName,
				name:     name,
				id:       id,
			})
		}
	}

	// Sort results for consistent output
	sortPackageDiffs := func(diffs []packageDiff) {
		sort.Slice(diffs, func(i, j int) bool {
			if diffs[i].provider != diffs[j].provider {
				return diffs[i].provider < diffs[j].provider
			}
			return diffs[i].name < diffs[j].name
		})
	}
	sortPackageDiffs(result.additions)
	sortPackageDiffs(result.removals)
	sortPackageDiffs(result.upgrades)

	return result
}

// displayDiff prints the diff results to stdout
func displayDiff(result diffResult) {
	hasChanges := len(result.additions) > 0 || len(result.removals) > 0 || len(result.upgrades) > 0

	if !hasChanges {
		fmt.Println("No drift detected. System is in sync with desired state.")
		return
	}

	fmt.Println("Drift detected between desired state and current state:")
	fmt.Println()

	if len(result.additions) > 0 {
		fmt.Println("Additions (packages in profiles but not installed):")
		for _, pkg := range result.additions {
			fmt.Printf("  + [%s] %s\n", pkg.provider, pkg.name)
		}
		fmt.Println()
	}

	if len(result.upgrades) > 0 {
		fmt.Println("Upgrades (packages with available updates):")
		for _, pkg := range result.upgrades {
			fmt.Printf("  ~ [%s] %s (upgrade available)\n", pkg.provider, pkg.name)
		}
		fmt.Println()
	}

	if len(result.removals) > 0 {
		fmt.Println("Removals (packages installed but not in any profile):")
		for _, pkg := range result.removals {
			fmt.Printf("  - [%s] %s\n", pkg.provider, pkg.name)
		}
		fmt.Println()
	}
}
