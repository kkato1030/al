package cmd

import (
	"fmt"
	"os"

	"github.com/kkato1030/al/internal/config"
	"github.com/kkato1030/al/internal/provider"
	"github.com/spf13/cobra"
)

// NewDiffCmd creates the diff command
func NewDiffCmd() *cobra.Command {
	var all bool
	var profile string

	cmd := &cobra.Command{
		Use:   "diff",
		Short: "Show differences between desired state (profiles) and current state (system)",
		Long: `Compare installed packages and configuration with the desired state defined in profiles.
Shows:
  + packages that need to be installed
  ~ packages that have upgrades available
  - packages that are installed but not in any profile

Defaults to checking the "core" profile. Use --all to check all AutoSync-enabled profiles, 
or --profile <name> to check a specific profile.
Exit code is non-zero when drift exists.`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if all && profile != "" {
				return fmt.Errorf("cannot use --all and --profile together")
			}
			return runDiff(all, profile)
		},
	}

	cmd.Flags().BoolVar(&all, "all", false, "Check all profiles with AutoSync enabled")
	cmd.Flags().StringVar(&profile, "profile", "", "Check only this profile and its extends")
	cmd.Flags().StringVar(&profile, "prf", "", "Short form of --profile")

	return cmd
}

func runDiff(all bool, profileName string) error {
	// Load profiles config
	profilesCfg, err := config.LoadProfilesConfig()
	if err != nil {
		return fmt.Errorf("failed to load profiles: %w", err)
	}

	// Determine target profiles
	var syncTargetNames []string
	if profileName != "" {
		// Specific profile requested - resolve profile and all its extends
		extends, err := config.ResolveProfileWithExtends(profileName)
		if err != nil {
			return fmt.Errorf("failed to resolve profile extends: %w", err)
		}
		syncTargetNames = extends
	} else if all {
		// All AutoSync profiles
		for _, p := range profilesCfg.Profiles {
			if p.AutoSync != nil && *p.AutoSync {
				syncTargetNames = append(syncTargetNames, p.Name)
			}
		}
		if len(syncTargetNames) == 0 {
			return fmt.Errorf("no profiles have AutoSync enabled")
		}
	} else {
		// Default: check "core" profile
		extends, err := config.ResolveProfileWithExtends("core")
		if err != nil {
			return fmt.Errorf("failed to resolve core profile: %w. Use --profile or --all flag", err)
		}
		syncTargetNames = extends
	}

	// Build set for quick lookup
	syncTargetSet := make(map[string]bool)
	for _, name := range syncTargetNames {
		syncTargetSet[name] = true
	}

	fmt.Printf("Checking profiles: %v\n", syncTargetNames)

	// Load packages config
	packagesCfg, err := config.LoadPackagesConfig()
	if err != nil {
		return fmt.Errorf("failed to load packages: %w", err)
	}

	// Track packages by status
	var toInstall []config.PackageConfig
	var toUpgrade []config.PackageConfig
	var alreadyInstalled []config.PackageConfig
	var manualPackages []config.PackageConfig

	// Cache provider status and package lists per provider
	type providerCache struct {
		provider       provider.Provider
		isInstalled    bool
		installedPkgs  map[string]bool
		upgradablePkgs map[string]bool
	}
	providerCaches := make(map[string]*providerCache)

	// First pass: collect packages by provider
	packagesByProvider := make(map[string][]config.PackageConfig)
	for _, pkg := range packagesCfg.Packages {
		if !syncTargetSet[pkg.Profile] {
			continue
		}
		if pkg.Provider == "manual" {
			manualPackages = append(manualPackages, pkg)
			continue
		}
		packagesByProvider[pkg.Provider] = append(packagesByProvider[pkg.Provider], pkg)
	}

	// Second pass: check each provider once and build caches
	for providerName := range packagesByProvider {
		debugLog("Processing provider: %s", providerName)

		p := getProvider(providerName)
		if p == nil {
			debugLog("Provider %s not available", providerName)
			continue
		}

		cache := &providerCache{provider: p}

		// Check if provider is installed
		providerInstalled, err := p.CheckInstalled()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to check if provider %s is installed: %v\n", providerName, err)
			cache.isInstalled = false
		} else {
			cache.isInstalled = providerInstalled
			debugLog("Provider %s installed: %v", providerName, providerInstalled)
		}

		// If provider is installed, get all installed and upgradable packages
		if cache.isInstalled {
			installedPkgs, err := p.ListInstalled()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to list installed packages for %s: %v\n", providerName, err)
				cache.installedPkgs = make(map[string]bool)
			} else {
				cache.installedPkgs = installedPkgs
			}

			upgradablePkgs, err := p.ListUpgradable()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to list upgradable packages for %s: %v\n", providerName, err)
				cache.upgradablePkgs = make(map[string]bool)
			} else {
				cache.upgradablePkgs = upgradablePkgs
			}
		}

		providerCaches[providerName] = cache
	}

	// Third pass: categorize each package
	for providerName, packages := range packagesByProvider {
		cache, ok := providerCaches[providerName]
		if !ok || cache.provider == nil {
			// Provider not available, mark all packages as needing install
			toInstall = append(toInstall, packages...)
			continue
		}

		if !cache.isInstalled {
			// Provider not installed, all packages need install
			toInstall = append(toInstall, packages...)
			continue
		}

		for _, pkg := range packages {
			if cache.installedPkgs[pkg.ID] {
				// Package is installed
				if cache.upgradablePkgs[pkg.ID] {
					toUpgrade = append(toUpgrade, pkg)
				} else {
					alreadyInstalled = append(alreadyInstalled, pkg)
				}
			} else {
				// Package not installed
				toInstall = append(toInstall, pkg)
			}
		}
	}

	// Fourth pass: find packages installed but not in ANY profile (removals)
	// Build a map of ALL packages across ALL profiles to avoid showing packages from other profiles
	allPackagesInConfig := make(map[string]bool)
	for _, pkg := range packagesCfg.Packages {
		if pkg.Provider != "manual" {
			allPackagesInConfig[pkg.ID] = true
		}
	}

	var toRemove []string
	for providerName, cache := range providerCaches {
		if !cache.isInstalled {
			continue
		}

		// Find installed packages not in ANY profile
		for installedPkgID := range cache.installedPkgs {
			if !allPackagesInConfig[installedPkgID] {
				toRemove = append(toRemove, fmt.Sprintf("%s %s", providerName, installedPkgID))
			}
		}
	}

	// Output results
	fmt.Println("\nDiff:")
	hasDrift := false

	// Show packages to install
	if len(toInstall) > 0 {
		hasDrift = true
		for _, pkg := range toInstall {
			fmt.Printf("  + %s %s\n", pkg.Provider, pkg.Name)
		}
	}

	// Show packages to upgrade
	if len(toUpgrade) > 0 {
		hasDrift = true
		for _, pkg := range toUpgrade {
			fmt.Printf("  ~ %s %s (upgrade available)\n", pkg.Provider, pkg.Name)
		}
	}

	// Show packages to remove (not in any profile)
	if len(toRemove) > 0 {
		hasDrift = true
		for _, pkgInfo := range toRemove {
			fmt.Printf("  - %s (not in any profile)\n", pkgInfo)
		}
	}

	// Show manual packages (informational)
	if len(manualPackages) > 0 {
		fmt.Println("\n  Manual packages (tracked but not managed):")
		for _, pkg := range manualPackages {
			fmt.Printf("    • %s\n", pkg.Name)
		}
	}

	// Summary
	if !hasDrift {
		fmt.Println("  No drift detected. System is in sync with profiles.")
		return nil
	}

	fmt.Printf("\nSummary:\n")
	fmt.Printf("  %d package(s) to install\n", len(toInstall))
	fmt.Printf("  %d package(s) to upgrade\n", len(toUpgrade))
	fmt.Printf("  %d package(s) not in any profile\n", len(toRemove))
	fmt.Printf("  %d package(s) already in sync\n", len(alreadyInstalled))

	// Exit with non-zero code when drift exists
	os.Exit(1)
	return nil
}
