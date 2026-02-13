package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kkato1030/al/internal/config"
	"github.com/kkato1030/al/internal/logger"
	"github.com/kkato1030/al/internal/provider"
	"github.com/kkato1030/al/internal/source"
	"github.com/kkato1030/al/internal/ui"
	"github.com/spf13/cobra"
)

// SyncPlanJSONOutput represents the JSON output format for sync --plan command
type SyncPlanJSONOutput struct {
	Actions []SyncAction `json:"actions"`
	Summary SyncSummary  `json:"summary"`
}

// SyncAction represents a single action in the sync plan
type SyncAction struct {
	Type     string `json:"type"`
	Provider string `json:"provider"`
	Name     string `json:"name"`
	ID       string `json:"id,omitempty"`
}

// SyncSummary provides summary counts for the sync plan
type SyncSummary struct {
	Providers      int      `json:"providers"`
	Installations  int      `json:"installations"`
	Upgrades       int      `json:"upgrades"`
	ManualPackages int      `json:"manual_packages"`
	Links          int      `json:"links"`
	Bootstrap      bool     `json:"bootstrap"`
	Profiles       []string `json:"profiles"`
}

// debugLog prints debug messages if AL_DEBUG environment variable is set
func debugLog(format string, args ...interface{}) {
	if os.Getenv("AL_DEBUG") != "" {
		timestamp := time.Now().Format("15:04:05.000")
		fmt.Fprintf(os.Stderr, "[DEBUG %s] ", timestamp)
		fmt.Fprintf(os.Stderr, format+"\n", args...)
	}
}

// loggedPrintf prints to stdout and optionally to the logger
func loggedPrintf(log *logger.Logger, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Print(msg)
	if log != nil {
		log.WriteString(msg)
	}
}

// loggedFprintf prints to the specified writer and optionally to the logger
func loggedFprintf(w io.Writer, log *logger.Logger, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprint(w, msg)
	if log != nil {
		log.WriteString(msg)
	}
}

// NewSyncCmd creates the sync command
func NewSyncCmd() *cobra.Command {
	var dryRun bool
	var plan bool
	var all bool
	var profile string
	var private bool
	var pkgOnly bool
	var linkOnly bool

	cmd := &cobra.Command{
		Use:   "sync [owner/repo]",
		Short: "Sync al environment: clone ~/.al if needed, then apply providers, packages, and links",
		Long:  "If AL_HOME (~/.al) does not exist, clones the given GitHub repository (owner/repo) into it, then applies. Otherwise applies only: ensures providers, installs packages in sync target profiles, applies link.d symlinks. Use --all to sync all AutoSync-enabled profiles, or --profile <name> to sync a specific profile and its extends. Use --pkg-only to sync only packages (skip links). Use --link-only to sync only links (skip packages). Use --private to clone a private repo: installs git, gh, git-credential-manager and runs gh auth login before cloning. Use --plan to preview changes without applying them.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateSyncFlags(all, profile, pkgOnly, linkOnly); err != nil {
				return err
			}
			// --plan takes precedence over --dry-run
			isPlanMode := plan || dryRun

			// Create logger for sync operation (skip in plan mode)
			var log *logger.Logger
			if !isPlanMode {
				configDir, err := config.GetConfigDir()
				if err == nil {
					logsDir := logger.GetLogsDir(configDir)
					cmdStr := "al sync"
					if len(args) > 0 {
						cmdStr += " " + strings.Join(args, " ")
					}
					log, err = logger.New(logsDir, cmdStr)
					if err != nil {
						// Log creation failed, but don't fail the sync
						fmt.Fprintf(os.Stderr, "Warning: failed to create log file: %v\n", err)
					}
				}
			}

			err := runSync(isPlanMode, all, profile, private, pkgOnly, linkOnly, args, log)

			if log != nil {
				log.Close()
			}

			return err
		},
	}

	cmd.Flags().BoolVar(&plan, "plan", false, "Preview what would be done without making changes")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be done without making changes (alias for --plan)")
	cmd.Flags().BoolVar(&all, "all", false, "Sync all profiles with AutoSync enabled")
	cmd.Flags().StringVar(&profile, "profile", "", "Sync only this profile and its extends (overrides AutoSync)")
	cmd.Flags().StringVar(&profile, "prf", "", "Short form of --profile")
	cmd.Flags().BoolVar(&pkgOnly, "pkg-only", false, "Sync only packages (skip links)")
	cmd.Flags().BoolVar(&linkOnly, "link-only", false, "Sync only links (skip packages)")
	cmd.Flags().BoolVar(&private, "private", false, "Clone a private repo: install git, gh, git-credential-manager and run gh auth login before cloning")

	return cmd
}

func validateSyncFlags(all bool, profile string, pkgOnly bool, linkOnly bool) error {
	if all && profile != "" {
		return fmt.Errorf("cannot use --all and --profile together")
	}
	if pkgOnly && linkOnly {
		return fmt.Errorf("cannot use --pkg-only and --link-only together")
	}
	return nil
}

func runSync(dryRun, all bool, profileName string, usePrivate bool, pkgOnly bool, linkOnly bool, args []string, log *logger.Logger) error {
	if log != nil {
		log.WriteString(fmt.Sprintf("Sync started (dryRun=%v, all=%v, profile=%s, pkgOnly=%v, linkOnly=%v)\n",
			dryRun, all, profileName, pkgOnly, linkOnly))
	}

	configDir, err := config.GetConfigDir()
	if err != nil {
		return fmt.Errorf("failed to get config directory: %w", err)
	}

	// AL_HOME does not exist → clone first
	if !pathExists(configDir) {
		var ownerRepo string
		if len(args) >= 1 && strings.TrimSpace(args[0]) != "" {
			ownerRepo = strings.TrimSpace(args[0])
		} else {
			ownerRepo, err = source.DefaultOwnerRepo()
			if err != nil {
				ownerRepo, err = promptOwnerRepo()
				if err != nil {
					return err
				}
			}
			if ownerRepo == "" {
				return fmt.Errorf("owner/repo is required to clone (e.g. kkato1030/dotal). Install and run 'gh auth login' to use default {account}/dotal")
			}
		}
		owner, repo, err := source.ParseOwnerRepo(ownerRepo)
		if err != nil {
			return err
		}
		if dryRun {
			if usePrivate {
				fmt.Printf("[dry-run] Would ensure git, gh, git-credential-manager and run gh auth login, then clone https://github.com/%s/%s into %s\n", owner, repo, configDir)
			} else {
				fmt.Printf("[dry-run] Would clone https://github.com/%s/%s into %s\n", owner, repo, configDir)
			}
			return nil
		}
		if usePrivate {
			if err := source.EnsurePrivateCloneSetup(); err != nil {
				return fmt.Errorf("private clone setup: %w", err)
			}
		}
		fmt.Printf("Cloning https://github.com/%s/%s into %s ...\n", owner, repo, configDir)
		if log != nil {
			log.WriteString(fmt.Sprintf("Cloning https://github.com/%s/%s into %s ...\n", owner, repo, configDir))
		}
		if err := source.Clone(configDir, owner, repo); err != nil {
			return fmt.Errorf("clone failed: %w", err)
		}
		if log != nil {
			log.WriteString("Clone completed successfully\n")
		}
	}

	doPackages := !linkOnly
	doLinks := !pkgOnly

	// Resolve sync target profiles (only needed for package sync)
	var syncTargetNames []string
	var syncTargetSet map[string]bool
	if doPackages {
		mode := "default"
		if profileName != "" {
			mode = "profile"
		} else if all {
			mode = "all"
		}
		var err error
		syncTargetNames, err = config.GetSyncTargetProfileNames(mode, profileName)
		if err != nil {
			return fmt.Errorf("sync target profiles: %w", err)
		}
		syncTargetSet = make(map[string]bool)
		for _, n := range syncTargetNames {
			syncTargetSet[n] = true
		}
	}

	if dryRun {
		debugLog("Starting plan mode")

		// Initialize variables for plan output
		var toInstall []config.PackageConfig
		var toUpgrade []config.PackageConfig
		var manualPackages []config.PackageConfig
		var providersToInstall []string

		if doPackages {
			debugLog("Loading packages configuration")
			if !IsJSONOutput() {
				fmt.Printf("\nPlan: Sync target profiles: %v\n", syncTargetNames)
			}
			packagesCfg, err := config.LoadPackagesConfig()
			if err != nil {
				return fmt.Errorf("load packages: %w", err)
			}
			debugLog("Loaded %d packages from config", len(packagesCfg.Packages))

			// Cache provider status and package lists per provider
			type providerCache struct {
				provider       provider.Provider
				isInstalled    bool
				installedPkgs  map[string]bool
				upgradablePkgs map[string]bool
			}
			providerCaches := make(map[string]*providerCache)

			// First pass: collect packages by provider
			debugLog("Collecting packages by provider")
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
			debugLog("Packages by provider: %v", func() map[string]int {
				counts := make(map[string]int)
				for p, pkgs := range packagesByProvider {
					counts[p] = len(pkgs)
				}
				return counts
			}())

			// Second pass: check each provider once and build caches
			debugLog("Checking providers and building caches")
			for providerName := range packagesByProvider {
				debugLog("Processing provider: %s", providerName)
				startTime := time.Now()

				p := getProvider(providerName)
				if p == nil {
					debugLog("Provider %s not available", providerName)
					continue
				}

				cache := &providerCache{provider: p}

				// Check if provider is installed
				debugLog("Checking if provider %s is installed", providerName)
				checkStart := time.Now()
				providerInstalled, err := p.CheckInstalled()
				debugLog("Provider %s CheckInstalled took %v", providerName, time.Since(checkStart))
				if err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to check if provider %s is installed: %v\n", providerName, err)
					cache.isInstalled = false
				} else {
					cache.isInstalled = providerInstalled
					debugLog("Provider %s installed: %v", providerName, providerInstalled)
				}

				// If provider is installed, get all installed and upgradable packages
				if cache.isInstalled {
					debugLog("Listing installed packages for %s", providerName)
					listStart := time.Now()
					installedPkgs, err := p.ListInstalled()
					debugLog("Provider %s ListInstalled took %v (returned %d packages)", providerName, time.Since(listStart), len(installedPkgs))
					if err != nil {
						fmt.Fprintf(os.Stderr, "Warning: failed to list installed packages for %s: %v\n", providerName, err)
						cache.installedPkgs = make(map[string]bool)
					} else {
						cache.installedPkgs = installedPkgs
					}

					debugLog("Listing upgradable packages for %s", providerName)
					upgradeStart := time.Now()
					upgradablePkgs, err := p.ListUpgradable()
					debugLog("Provider %s ListUpgradable took %v (returned %d packages)", providerName, time.Since(upgradeStart), len(upgradablePkgs))
					if err != nil {
						fmt.Fprintf(os.Stderr, "Warning: failed to list upgradable packages for %s: %v\n", providerName, err)
						cache.upgradablePkgs = make(map[string]bool)
					} else {
						cache.upgradablePkgs = upgradablePkgs
					}
				} else {
					debugLog("Provider %s not installed, skipping package checks", providerName)
				}

				providerCaches[providerName] = cache
				debugLog("Finished processing provider %s in %v", providerName, time.Since(startTime))
			}

			// Third pass: categorize each package
			debugLog("Categorizing packages")
			for providerName, packages := range packagesByProvider {
				cache, ok := providerCaches[providerName]
				if !ok || cache.provider == nil {
					// Provider not available, mark all packages for install
					debugLog("Provider %s not available, marking %d packages for install", providerName, len(packages))
					toInstall = append(toInstall, packages...)
					continue
				}

				if !cache.isInstalled {
					// Provider not installed, all packages need install
					debugLog("Provider %s not installed, marking %d packages for install", providerName, len(packages))
					toInstall = append(toInstall, packages...)
					continue
				}

				for _, pkg := range packages {
					if cache.installedPkgs[pkg.ID] {
						// Package is installed, check if upgradable
						if cache.upgradablePkgs[pkg.ID] {
							debugLog("Package %s is upgradable", pkg.Name)
							toUpgrade = append(toUpgrade, pkg)
						} else {
							debugLog("Package %s is already installed and up-to-date", pkg.Name)
						}
					} else {
						// Package not installed
						debugLog("Package %s needs to be installed", pkg.Name)
						toInstall = append(toInstall, pkg)
					}
				}
			}

			debugLog("Categorization complete: %d to install, %d to upgrade, %d manual", len(toInstall), len(toUpgrade), len(manualPackages))

			// Collect providers to install
			var providersToInstall []string
			var providersNeeded []string
			for _, pkg := range packagesCfg.Packages {
				if syncTargetSet[pkg.Profile] && pkg.Provider != "manual" {
					providersNeeded = append(providersNeeded, pkg.Provider)
				}
			}
			orderedProviders, err := config.ResolveProvidersWithDependencies(providersNeeded)
			if err != nil {
				if !IsJSONOutput() {
					fmt.Fprintf(os.Stderr, "Warning: failed to resolve provider dependencies: %v\n", err)
				}
				orderedProviders = providersNeeded
			}
			for _, name := range orderedProviders {
				p := getProvider(name)
				if p == nil {
					continue
				}
				installed, err := p.CheckInstalled()
				if err != nil {
					if !IsJSONOutput() {
						fmt.Fprintf(os.Stderr, "Warning: failed to check if provider %s is installed: %v\n", name, err)
					}
					continue
				}
				if !installed {
					providersToInstall = append(providersToInstall, name)
				}
			}
		}

		// Collect links
		var links []config.LinkEntry
		if doLinks {
			var err error
			links, err = config.ListLinks("", "")
			if err != nil {
				if !IsJSONOutput() {
					fmt.Fprintf(os.Stderr, "Warning: failed to list links: %v\n", err)
				}
				links = []config.LinkEntry{}
			}
		}

		// Check for bootstrap script
		hasBootstrap := false
		exists, err := config.BootstrapScriptExists()
		if err != nil {
			if !IsJSONOutput() {
				fmt.Fprintf(os.Stderr, "Warning: failed to check for bootstrap script: %v\n", err)
			}
		} else {
			hasBootstrap = exists
		}

		// Output the plan
		outputSyncPlan(syncTargetNames, toInstall, toUpgrade, manualPackages, providersToInstall, links, hasBootstrap)
		return nil
	}

	if doPackages {
		packagesCfg, err := config.LoadPackagesConfig()
		if err != nil {
			return fmt.Errorf("load packages: %w", err)
		}
		var providersNeeded []string
		var manualPackages []config.PackageConfig
		for _, pkg := range packagesCfg.Packages {
			if syncTargetSet[pkg.Profile] {
				if pkg.Provider != "manual" {
					providersNeeded = append(providersNeeded, pkg.Provider)
				} else {
					manualPackages = append(manualPackages, pkg)
				}
			}
		}
		orderedProviders, err := config.ResolveProvidersWithDependencies(providersNeeded)
		if err != nil {
			return fmt.Errorf("resolve provider dependencies: %w", err)
		}
		for _, name := range orderedProviders {
			p := getProvider(name)
			if p == nil {
				continue
			}
			installed, err := p.CheckInstalled()
			if err != nil {
				return fmt.Errorf("check provider %s: %w", name, err)
			}
			if !installed {
				fmt.Printf("Installing provider %s...\n", name)
				if err := p.Install(); err != nil {
					return fmt.Errorf("install provider %s: %w", name, err)
				}
				if err := p.SetupConfig(); err != nil {
					return fmt.Errorf("setup provider %s: %w", name, err)
				}
			}
		}
		for _, pkg := range packagesCfg.Packages {
			if !syncTargetSet[pkg.Profile] {
				continue
			}
			p := getProvider(pkg.Provider)
			if p == nil {
				continue
			}
			installed, _ := p.CheckInstalled()
			if !installed {
				continue
			}
			fmt.Printf("Installing package %s (%s)...\n", pkg.Name, pkg.ID)
			if err := p.InstallPackage(pkg.ID); err != nil {
				return fmt.Errorf("install package %s: %w", pkg.Name, err)
			}
		}
		// Display warning for manual packages
		if len(manualPackages) > 0 {
			fmt.Fprintln(os.Stderr, "\n⚠️  Manual packages detected:")
			fmt.Fprintln(os.Stderr, "The following packages are tracked as manually installed.")
			fmt.Fprintln(os.Stderr, "Please ensure they are installed on this system:")
			for _, pkg := range manualPackages {
				fmt.Fprintf(os.Stderr, "  - %s (ID: %s, Profile: %s)\n", pkg.Name, pkg.ID, pkg.Profile)
			}
			fmt.Fprintln(os.Stderr)
		}
	}

	if doLinks {
		linkDir, err := config.GetLinkDir()
		if err != nil {
			return fmt.Errorf("get link dir: %w", err)
		}
		entries, err := config.ListLinks("", "")
		if err != nil {
			return fmt.Errorf("list links: %w", err)
		}
		for _, entry := range entries {
			entryDir := filepath.Join(linkDir, entry.Name)
			if err := config.EnsureLinkSymlink(&entry, entryDir); err != nil {
				return fmt.Errorf("link %s: %w", entry.Name, err)
			}
		}
	}

	// Run bootstrap script if present
	exists, err := config.BootstrapScriptExists()
	if err != nil {
		return fmt.Errorf("bootstrap check: %w", err)
	}
	if exists {
		scriptPath, err := config.GetBootstrapScriptPath()
		if err != nil {
			return fmt.Errorf("bootstrap path: %w", err)
		}
		bootstrapDir, err := config.GetBootstrapDir()
		if err != nil {
			return fmt.Errorf("bootstrap dir: %w", err)
		}
		fmt.Println("\nRunning bootstrap script...")
		cmd := exec.Command(scriptPath)
		cmd.Dir = bootstrapDir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("bootstrap script: %w", err)
		}
	}

	fmt.Println("\nSync complete. Add the following to your shell config (.zshrc, .bashrc, etc.):")
	fmt.Println(`  eval "$(al activate zsh)"  # or bash`)
	if log != nil {
		log.WriteString("\nSync completed successfully\n")
	}
	return nil
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func promptOwnerRepo() (string, error) {
	model := ui.NewTextInputModel("GitHub repository (owner/repo)", true)
	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		return "", fmt.Errorf("prompt failed: %w", err)
	}
	return strings.TrimSpace(model.GetValue()), nil
}

func getProvider(name string) provider.Provider {
	switch name {
	case "brew":
		return provider.NewBrewProvider()
	case "mas":
		return provider.NewMasProvider()
	case "manual":
		return provider.NewManualProvider()
	default:
		return nil
	}
}

// outputSyncPlan outputs the sync plan in either JSON or human-readable format
func outputSyncPlan(
	syncTargetNames []string,
	toInstall, toUpgrade, manualPackages []config.PackageConfig,
	providersToInstall []string,
	links []config.LinkEntry,
	hasBootstrap bool,
) {
	if IsJSONOutput() {
		// Build JSON output
		actions := make([]SyncAction, 0)

		// Add provider installations
		for _, name := range providersToInstall {
			actions = append(actions, SyncAction{
				Type:     "install_provider",
				Provider: name,
				Name:     name,
			})
		}

		// Add package installations
		for _, pkg := range toInstall {
			actions = append(actions, SyncAction{
				Type:     "install",
				Provider: pkg.Provider,
				Name:     pkg.Name,
				ID:       pkg.ID,
			})
		}

		// Add package upgrades
		for _, pkg := range toUpgrade {
			actions = append(actions, SyncAction{
				Type:     "upgrade",
				Provider: pkg.Provider,
				Name:     pkg.Name,
				ID:       pkg.ID,
			})
		}

		// Add manual packages
		for _, pkg := range manualPackages {
			actions = append(actions, SyncAction{
				Type:     "manual",
				Provider: pkg.Provider,
				Name:     pkg.Name,
				ID:       pkg.ID,
			})
		}

		// Add links
		for _, link := range links {
			actions = append(actions, SyncAction{
				Type: "link",
				Name: link.Manifest.UserPath,
			})
		}

		// Add bootstrap script
		if hasBootstrap {
			actions = append(actions, SyncAction{
				Type: "bootstrap",
				Name: "bootstrap script",
			})
		}

		output := SyncPlanJSONOutput{
			Actions: actions,
			Summary: SyncSummary{
				Providers:      len(providersToInstall),
				Installations:  len(toInstall),
				Upgrades:       len(toUpgrade),
				ManualPackages: len(manualPackages),
				Links:          len(links),
				Bootstrap:      hasBootstrap,
				Profiles:       syncTargetNames,
			},
		}

		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(output); err != nil {
			fmt.Fprintf(os.Stderr, "error encoding JSON: %v\n", err)
		}
		return
	}

	// Human-readable output
	fmt.Printf("\nPlan: Sync target profiles: %v\n", syncTargetNames)
	fmt.Println("\nPlan:")

	// Show providers to install
	for _, name := range providersToInstall {
		fmt.Printf("  + provider %s\n", name)
	}

	// Show packages to install
	if len(toInstall) > 0 {
		for _, pkg := range toInstall {
			fmt.Printf("  + %s %s\n", pkg.Provider, pkg.Name)
		}
	}

	// Show packages to upgrade
	if len(toUpgrade) > 0 {
		for _, pkg := range toUpgrade {
			fmt.Printf("  ~ %s %s (upgrade)\n", pkg.Provider, pkg.Name)
		}
	}

	// Show manual packages
	if len(manualPackages) > 0 {
		fmt.Println("\n  Manual packages (ensure they are installed):")
		for _, pkg := range manualPackages {
			fmt.Printf("    - %s\n", pkg.Name)
		}
	}

	// Show links
	if len(links) > 0 {
		fmt.Println("\n  Links:")
		for _, link := range links {
			fmt.Printf("  + link %s\n", link.Manifest.UserPath)
		}
	}

	// Show bootstrap script
	if hasBootstrap {
		fmt.Println("\n  + bootstrap script")
	}

	fmt.Println("\nNo changes will be made in plan mode.")
}
