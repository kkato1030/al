package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kkato1030/al/internal/config"
	"github.com/kkato1030/al/internal/provider"
	"github.com/kkato1030/al/internal/source"
	"github.com/kkato1030/al/internal/ui"
	"github.com/spf13/cobra"
)

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
			return runSync(isPlanMode, all, profile, private, pkgOnly, linkOnly, args)
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

func runSync(dryRun, all bool, profileName string, usePrivate bool, pkgOnly bool, linkOnly bool, args []string) error {
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
		if err := source.Clone(configDir, owner, repo); err != nil {
			return fmt.Errorf("clone failed: %w", err)
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
		if doPackages {
			fmt.Printf("\nPlan: Sync target profiles: %v\n", syncTargetNames)
			packagesCfg, err := config.LoadPackagesConfig()
			if err != nil {
				return fmt.Errorf("load packages: %w", err)
			}

			// Track packages by status
			var toInstall []config.PackageConfig
			var toUpgrade []config.PackageConfig
			var manualPackages []config.PackageConfig

			for _, pkg := range packagesCfg.Packages {
				if !syncTargetSet[pkg.Profile] {
					continue
				}

				if pkg.Provider == "manual" {
					manualPackages = append(manualPackages, pkg)
					continue
				}

				p := getProvider(pkg.Provider)
				if p == nil {
					continue
				}

				// Check if provider is installed
				providerInstalled, err := p.CheckInstalled()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to check if provider %s is installed: %v\n", pkg.Provider, err)
					toInstall = append(toInstall, pkg)
					continue
				}
				if !providerInstalled {
					toInstall = append(toInstall, pkg)
					continue
				}

				// Check package status
				pkgInstalled, err := p.IsPackageInstalled(pkg.ID)
				if err != nil {
					// If we can't check, assume it needs install
					toInstall = append(toInstall, pkg)
					continue
				}

				if !pkgInstalled {
					toInstall = append(toInstall, pkg)
				} else {
					// Check if upgradable
					upgradable, err := p.IsPackageUpgradable(pkg.ID)
					if err == nil && upgradable {
						toUpgrade = append(toUpgrade, pkg)
					}
				}
			}

			fmt.Println("\nPlan:")

			// Show providers to install
			var providersNeeded []string
			for _, pkg := range packagesCfg.Packages {
				if syncTargetSet[pkg.Profile] && pkg.Provider != "manual" {
					providersNeeded = append(providersNeeded, pkg.Provider)
				}
			}
			orderedProviders, err := config.ResolveProvidersWithDependencies(providersNeeded)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to resolve provider dependencies: %v\n", err)
				orderedProviders = providersNeeded
			}
			for _, name := range orderedProviders {
				p := getProvider(name)
				if p == nil {
					continue
				}
				installed, err := p.CheckInstalled()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to check if provider %s is installed: %v\n", name, err)
					continue
				}
				if !installed {
					fmt.Printf("  + provider %s\n", name)
				}
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
		}

		if doLinks {
			links, err := config.ListLinks("", "")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to list links: %v\n", err)
			} else if len(links) > 0 {
				fmt.Println("\n  Links:")
				for _, link := range links {
					fmt.Printf("  + link %s\n", link.Manifest.UserPath)
				}
			}
		}

		exists, err := config.BootstrapScriptExists()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to check for bootstrap script: %v\n", err)
		} else if exists {
			fmt.Println("\n  + bootstrap script")
		}

		fmt.Println("\nNo changes will be made in plan mode.")
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
