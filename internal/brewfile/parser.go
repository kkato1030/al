package brewfile

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// Entry represents a single parsed Brewfile entry (tap, brew, cask, or mas).
type Entry struct {
	Provider string // "brew" or "mas"
	ID       string // e.g. "formula:ruby", "cask:firefox", "tap:user/repo", "1234567890"
	Name     string // display name (for mas: app name; for brew: same as package part of ID)
}

// SkippedLine records a line that was skipped (unsupported or parse error).
type SkippedLine struct {
	LineNum int
	Line    string
	Reason  string
}

// ParseResult holds the result of parsing a Brewfile.
type ParseResult struct {
	Entries []Entry
	Skipped []SkippedLine
}

// Line patterns (simplified; we don't run Ruby).
// tap "user/repo" or tap 'user/repo'
var tapRegex = regexp.MustCompile(`^\s*tap\s+["']([^"']+)["']`)

// brew "formula" or brew "formula@16" (optional , ...)
var brewRegex = regexp.MustCompile(`^\s*brew\s+["']([^"']+)["']`)

// cask "name"
var caskRegex = regexp.MustCompile(`^\s*cask\s+["']([^"']+)["']`)

// mas "App Name", id: 1234567890
var masRegex = regexp.MustCompile(`^\s*mas\s+["']([^"']+)["']\s*,?\s*id\s*:\s*(\d+)`)

// Unsupported keywords (skip with reason)
var unsupportedPrefixes = []struct {
	prefix string
	label  string
}{
	{"vscode", "vscode"},
	{"go ", "go"},
	{"cargo ", "cargo"},
	{"flatpak", "flatpak"},
	{"cask_args", "cask_args"},
}

// ParseFile reads path and parses the Brewfile, returning entries and skipped lines.
func ParseFile(path string) (*ParseResult, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var entries []Entry
	var skipped []SkippedLine
	scanner := bufio.NewScanner(f)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		entry, skipReason := parseLine(trimmed)
		if skipReason != "" {
			skipped = append(skipped, SkippedLine{LineNum: lineNum, Line: line, Reason: skipReason})
			continue
		}
		if entry != nil {
			entries = append(entries, *entry)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return &ParseResult{Entries: entries, Skipped: skipped}, nil
}

// parseLine parses a single non-empty, non-comment line.
// Returns (entry, "") if parsed, (nil, reason) if skipped, (nil, "") for unknown/unsupported.
func parseLine(line string) (*Entry, string) {
	// tap "user/repo" — only single-arg tap
	if m := tapRegex.FindStringSubmatch(line); len(m) == 2 {
		tapName := m[1]
		if strings.Contains(tapName, " ") {
			return nil, "tap with multiple args"
		}
		return &Entry{Provider: "brew", ID: "tap:" + tapName, Name: tapName}, ""
	}
	if strings.TrimSpace(line) != "" && (strings.HasPrefix(strings.TrimSpace(line), "tap ") || strings.HasPrefix(strings.TrimSpace(line), "tap\t")) {
		// tap with URL etc
		return nil, "tap (unsupported format)"
	}

	// brew "formula"
	if m := brewRegex.FindStringSubmatch(line); len(m) == 2 {
		name := m[1]
		return &Entry{Provider: "brew", ID: "formula:" + name, Name: name}, ""
	}

	// cask "name"
	if m := caskRegex.FindStringSubmatch(line); len(m) == 2 {
		name := m[1]
		return &Entry{Provider: "brew", ID: "cask:" + name, Name: name}, ""
	}

	// mas "App Name", id: 1234567890
	if m := masRegex.FindStringSubmatch(line); len(m) == 3 {
		appName := m[1]
		appID := m[2]
		return &Entry{Provider: "mas", ID: appID, Name: appName}, ""
	}
	if strings.TrimSpace(line) != "" && (strings.HasPrefix(strings.TrimSpace(line), "mas ") || strings.HasPrefix(strings.TrimSpace(line), "mas\t")) {
		return nil, "mas (missing or invalid id)"
	}

	// Unsupported types
	lower := strings.ToLower(strings.TrimSpace(line))
	for _, u := range unsupportedPrefixes {
		if strings.HasPrefix(lower, u.prefix) {
			return nil, u.label
		}
	}

	// Unknown
	if looksLikeRubyLine(line) {
		return nil, "unsupported"
	}
	return nil, ""
}

func looksLikeRubyLine(line string) bool {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return false
	}
	// Common Brewfile directives
	for _, prefix := range []string{"tap ", "brew ", "cask ", "mas ", "vscode ", "cask_args", "tap_args"} {
		if strings.HasPrefix(trimmed, prefix) || strings.HasPrefix(trimmed, strings.TrimRight(prefix, " ")) {
			return true
		}
	}
	return false
}

// ResolveBrewfilePath returns the path to use for the Brewfile.
// If userPath is non-empty, it is returned as-is (caller should check existence).
// If empty, tries ./Brewfile then ~/.Brewfile.
func ResolveBrewfilePath(userPath string) (string, error) {
	if userPath != "" {
		return userPath, nil
	}
	// Default: ./Brewfile
	if _, err := os.Stat("Brewfile"); err == nil {
		return "Brewfile", nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home dir: %w", err)
	}
	defaultPath := home + "/.Brewfile"
	if _, err := os.Stat(defaultPath); err == nil {
		return defaultPath, nil
	}
	return "Brewfile", nil // return Brewfile so that os.Open gives a clear "no such file" error
}
