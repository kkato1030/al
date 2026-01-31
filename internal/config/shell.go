package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const shellManifestFilename = ".manifest.json"

// ShellManifest represents the manifest for a package's shell.d directory.
// It holds load order (after) and whether this package's shell is enabled for `al activate`.
type ShellManifest struct {
	After   string `json:"after,omitempty"`   // package dir name that this should load after
	Enabled bool   `json:"enabled"`           // whether to source in `al activate`
}

// GetShellDir returns the path to ~/.al/shell.d/
func GetShellDir() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "shell.d"), nil
}

// PackageDirName returns a safe directory name for a package (id, provider).
// Used as the subdirectory name under shell.d. Profile is not included so the same package shares one shell.d across profiles.
func PackageDirName(id, provider string) string {
	sanitize := func(s string) string {
		return strings.NewReplacer("/", "_", ":", "_", " ", "_").Replace(s)
	}
	return sanitize(id) + "_" + sanitize(provider)
}

// GetShellPackageDir returns the path to ~/.al/shell.d/<PackageDirName(id,provider)>/
func GetShellPackageDir(id, provider string) (string, error) {
	shellDir, err := GetShellDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(shellDir, PackageDirName(id, provider)), nil
}

// EnsureShellPackageDir creates the package's shell.d directory if it doesn't exist.
func EnsureShellPackageDir(id, provider string) error {
	if err := EnsureConfigDir(); err != nil {
		return err
	}
	shellDir, err := GetShellDir()
	if err != nil {
		return err
	}
	pkgDir := filepath.Join(shellDir, PackageDirName(id, provider))
	return os.MkdirAll(pkgDir, 0755)
}

// RemoveShellPackageDir removes the package's shell.d directory and all its contents.
// It is a no-op if the directory does not exist.
func RemoveShellPackageDir(id, provider string) error {
	pkgDir, err := GetShellPackageDir(id, provider)
	if err != nil {
		return err
	}
	if _, err := os.Stat(pkgDir); os.IsNotExist(err) {
		return nil
	}
	return os.RemoveAll(pkgDir)
}

// LoadShellManifest loads the manifest from a package's shell.d directory.
// If the file does not exist, returns a default manifest (Enabled: true).
func LoadShellManifest(pkgDir string) (*ShellManifest, error) {
	manifestPath := filepath.Join(pkgDir, shellManifestFilename)
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		return &ShellManifest{Enabled: true}, nil
	}
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, err
	}
	var m ShellManifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

// SaveShellManifest saves the manifest to a package's shell.d directory.
func SaveShellManifest(pkgDir string, m *ShellManifest) error {
	if err := os.MkdirAll(pkgDir, 0755); err != nil {
		return err
	}
	manifestPath := filepath.Join(pkgDir, shellManifestFilename)
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(manifestPath, data, 0644)
}

// ShellEntry represents one package's shell.d entry: its directory name and snippet file paths (for a given shell ext).
type ShellEntry struct {
	DirName string   // directory name under shell.d (PackageDirName)
	Paths   []string // absolute paths to snippet files (e.g. .zsh) in load order (same rank = file name dict order)
}

// ListShellPackageDirNames returns all directory names under shell.d (each is a package's shell dir).
func ListShellPackageDirNames() ([]string, error) {
	shellDir, err := GetShellDir()
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(shellDir); os.IsNotExist(err) {
		return nil, nil
	}
	entries, err := os.ReadDir(shellDir)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
			names = append(names, e.Name())
		}
	}
	return names, nil
}

// SnippetFilesInDir returns snippet file paths in pkgDir with the given extension (e.g. ".zsh"), in dictionary order.
func SnippetFilesInDir(pkgDir, ext string) ([]string, error) {
	entries, err := os.ReadDir(pkgDir)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.HasSuffix(e.Name(), ext) && e.Name() != shellManifestFilename {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	paths := make([]string, len(names))
	for i, n := range names {
		paths[i] = filepath.Join(pkgDir, n)
	}
	return paths, nil
}

// GetEnabledShellEntriesInOrder returns enabled shell entries in topological order (after dependency).
// Ties are broken by directory name dictionary order. Only entries with at least one snippet file for ext are included.
func GetEnabledShellEntriesInOrder(ext string) ([]ShellEntry, error) {
	shellDir, err := GetShellDir()
	if err != nil {
		return nil, err
	}
	dirNames, err := ListShellPackageDirNames()
	if err != nil {
		return nil, err
	}

	// Build: dirName -> manifest, and collect enabled dirs with their snippet paths
	type node struct {
		after string
		paths []string
	}
	nodes := make(map[string]*node)
	var enabledDirs []string
	for _, dirName := range dirNames {
		pkgDir := filepath.Join(shellDir, dirName)
		manifest, err := LoadShellManifest(pkgDir)
		if err != nil {
			return nil, err
		}
		if !manifest.Enabled {
			continue
		}
		paths, err := SnippetFilesInDir(pkgDir, ext)
		if err != nil {
			return nil, err
		}
		if len(paths) == 0 {
			continue
		}
		enabledDirs = append(enabledDirs, dirName)
		nodes[dirName] = &node{after: manifest.After, paths: paths}
	}

	dirSet := make(map[string]bool)
	for _, d := range enabledDirs {
		dirSet[d] = true
	}
	afterOf := make(map[string]string)
	for d, n := range nodes {
		afterOf[d] = n.after
	}
	// Topological sort: "after" means "this loads after after", so edge is after -> dirName.
	// Only count dependency if "after" is in the enabled set.
	order, err := topologicalSortShell(enabledDirs, afterOf, dirSet)
	if err != nil {
		return nil, err
	}

	result := make([]ShellEntry, 0, len(order))
	for _, dirName := range order {
		result = append(result, ShellEntry{DirName: dirName, Paths: nodes[dirName].paths})
	}
	return result, nil
}

// topologicalSortShell orders dirNames so that for each node, any "after" target appears before it.
// Ties (same rank) are broken by directory name dictionary order.
// dirSet is the set of enabled directory names; dependencies on names not in dirSet are ignored.
func topologicalSortShell(dirNames []string, afterOf map[string]string, dirSet map[string]bool) ([]string, error) {
	inDegree := make(map[string]int)
	for _, d := range dirNames {
		inDegree[d] = 0
	}
	for _, d := range dirNames {
		after := afterOf[d]
		if after != "" && dirSet[after] {
			inDegree[d]++
		}
	}
	var queue []string
	for _, d := range dirNames {
		if inDegree[d] == 0 {
			queue = append(queue, d)
		}
	}
	sort.Strings(queue)
	var order []string
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		order = append(order, cur)
		var next []string
		for _, d := range dirNames {
			if afterOf[d] == cur {
				inDegree[d]--
				if inDegree[d] == 0 {
					next = append(next, d)
				}
			}
		}
		sort.Strings(next)
		queue = append(queue, next...)
		sort.Strings(queue)
	}
	if len(order) != len(dirNames) {
		return nil, fmt.Errorf("shell.d: cycle in --after dependency")
	}
	return order, nil
}
