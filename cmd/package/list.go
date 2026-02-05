package packagecmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/kkato1030/al/internal/config"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

const (
	profileLineChar   = "━"  // U+2501 BOX DRAWINGS HEAVY HORIZONTAL
	profileLineLead   = "━━ " // leading "━━ " before profile name
	profileLineTail   = " ━"  // " ━" after profile name before fill
	maxProfileLineLen = 48   // cap line length so it doesn't stretch across wide terminals
)

// NewPackageListCmd creates the package list command
func NewPackageListCmd() *cobra.Command {
	var profile string
	var provider string
	var noPager bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all packages",
		Long:  "List all configured packages. Optionally filter by profile and/or provider. Output is piped through a pager (e.g. less) when stdout is a TTY.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPackageListWithPager(profile, provider, noPager)
		},
	}

	cmd.Flags().StringVar(&profile, "profile", "", "Filter packages by profile name")
	cmd.Flags().StringVar(&profile, "prf", "", "Short form of --profile")
	cmd.Flags().StringVar(&provider, "provider", "", "Filter packages by provider name")
	cmd.Flags().StringVar(&provider, "prv", "", "Short form of --provider")
	cmd.Flags().BoolVar(&noPager, "no-pager", false, "Do not pipe output through a pager")

	return cmd
}

// terminalWidth returns the terminal width, or 80 if not available.
func terminalWidth() int {
	if w, _, err := term.GetSize(int(os.Stderr.Fd())); err == nil && w > 0 {
		return w
	}
	return 80
}

// profileSeparatorLine returns a short line like "━━ core ━━━━━━━━━━━━━" (capped at maxProfileLineLen so it stays readable).
func profileSeparatorLine(profileName string) string {
	prefix := profileLineLead + profileName + profileLineTail
	prefixRunes := utf8.RuneCountInString(prefix)
	fill := maxProfileLineLen - prefixRunes
	if fill < 0 {
		fill = 0
	}
	return prefix + strings.Repeat(profileLineChar, fill)
}

// formatPackageGrid formats package names in a grid with space padding. indent is prepended to each line; contentWidth is max line width for the grid (excluding indent); columnWidth is cell width (name + padding).
func formatPackageGrid(names []string, indent string, contentWidth, columnWidth int) []string {
	if len(names) == 0 {
		return nil
	}
	if columnWidth <= 0 {
		columnWidth = 1
	}
	numCols := contentWidth / columnWidth
	if numCols < 1 {
		numCols = 1
	}
	var lines []string
	for rowStart := 0; rowStart < len(names); rowStart += numCols {
		var row strings.Builder
		row.WriteString(indent)
		for c := 0; c < numCols; c++ {
			idx := rowStart + c
			if idx >= len(names) {
				break
			}
			name := names[idx]
			row.WriteString(name)
			written := utf8.RuneCountInString(name)
			pad := columnWidth - written
			if pad > 0 {
				row.WriteString(strings.Repeat(" ", pad))
			}
		}
		lines = append(lines, strings.TrimRight(row.String(), " "))
	}
	return lines
}

// runPackageListWithPager runs the list; when stdout is a TTY and noPager is false, pipes output through $PAGER (default: less -R).
// After the pager exits, the list is printed again to stdout so it remains visible in the terminal.
func runPackageListWithPager(profileFilter, providerFilter string, noPager bool) error {
	stdout := os.Stdout
	usePager := !noPager && term.IsTerminal(int(stdout.Fd()))
	if !usePager {
		return runPackageList(stdout, profileFilter, providerFilter)
	}
	var buf bytes.Buffer
	if err := runPackageList(&buf, profileFilter, providerFilter); err != nil {
		return err
	}
	// Keep a copy; the pager will consume buf when reading from stdin.
	saved := make([]byte, buf.Len())
	copy(saved, buf.Bytes())
	pagerCmd, pagerArgs := pagerCommand()
	cmd := exec.Command(pagerCmd, pagerArgs...)
	cmd.Stdin = &buf
	cmd.Stdout = stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	// Re-print the list so it stays visible in the terminal after the pager is closed.
	_, err := stdout.Write(saved)
	return err
}

func pagerCommand() (name string, args []string) {
	pager := os.Getenv("PAGER")
	if pager == "" {
		return "less", []string{"-R"}
	}
	parts := strings.Fields(pager)
	if len(parts) == 0 {
		return "less", []string{"-R"}
	}
	name = parts[0]
	args = parts[1:]
	if name == "less" && !contains(args, "-R") {
		args = append(args, "-R")
	}
	return name, args
}

func contains(s []string, x string) bool {
	for _, v := range s {
		if v == x {
			return true
		}
	}
	return false
}

func runPackageList(w io.Writer, profileFilter, providerFilter string) error {
	packagesConfig, err := config.LoadPackagesConfig()
	if err != nil {
		return fmt.Errorf("error loading packages config: %w", err)
	}

	// Filter packages based on provided flags
	var filteredPackages []config.PackageConfig
	for _, pkg := range packagesConfig.Packages {
		matchesProfile := profileFilter == "" || pkg.Profile == profileFilter
		matchesProvider := providerFilter == "" || pkg.Provider == providerFilter

		if matchesProfile && matchesProvider {
			filteredPackages = append(filteredPackages, pkg)
		}
	}

	if len(filteredPackages) == 0 {
		if profileFilter != "" || providerFilter != "" {
			fmt.Fprintln(w, "No packages found matching the specified filters")
		} else {
			fmt.Fprintln(w, "No packages configured")
		}
		return nil
	}

	// Group packages by profile and provider
	// Structure: map[profile]map[provider][]PackageConfig
	grouped := make(map[string]map[string][]config.PackageConfig)
	for _, pkg := range filteredPackages {
		profileName := pkg.Profile
		if profileName == "" {
			profileName = "(no profile)"
		}
		providerName := pkg.Provider
		if providerName == "" {
			providerName = "(no provider)"
		}

		if grouped[profileName] == nil {
			grouped[profileName] = make(map[string][]config.PackageConfig)
		}
		grouped[profileName][providerName] = append(grouped[profileName][providerName], pkg)
	}

	// Sort profiles and providers for consistent output
	profiles := make([]string, 0, len(grouped))
	for profile := range grouped {
		profiles = append(profiles, profile)
	}
	sort.Strings(profiles)

	width := terminalWidth()
	if width < 40 {
		width = 80
	}
	const packageIndent = "    " // 4 spaces (about half of typical tab width)
	const indentDisplayWidth = 4
	contentWidth := width - indentDisplayWidth
	if contentWidth < 20 {
		contentWidth = 72
	}

	fmt.Fprintln(w, "Configured packages")
	fmt.Fprintln(w)
	for _, profileName := range profiles {
		fmt.Fprintln(w, profileSeparatorLine(profileName))

		providers := make([]string, 0, len(grouped[profileName]))
		for provider := range grouped[profileName] {
			providers = append(providers, provider)
		}
		sort.Strings(providers)

		for _, providerName := range providers {
			packages := grouped[profileName][providerName]
			sort.Slice(packages, func(i, j int) bool {
				return packages[i].Name < packages[j].Name
			})

			packageNames := make([]string, len(packages))
			for idx, pkg := range packages {
				packageNames[idx] = pkg.Name
			}
			fmt.Fprintf(w, "[%s] (%d)\n", providerName, len(packageNames))

			maxNameLen := 0
			for _, name := range packageNames {
				n := utf8.RuneCountInString(name)
				if n > maxNameLen {
					maxNameLen = n
				}
			}
			columnWidth := maxNameLen + 1
			for _, line := range formatPackageGrid(packageNames, packageIndent, contentWidth, columnWidth) {
				fmt.Fprintln(w, line)
			}
			fmt.Fprintln(w)
		}
	}

	return nil
}
