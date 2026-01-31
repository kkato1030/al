package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const linkManifestFilename = ".manifest.json"
const linkContentName = "content"

// LinkType is either "file" or "dir".
type LinkType string

const (
	LinkTypeFile LinkType = "file"
	LinkTypeDir  LinkType = "dir"
)

// LinkManifest represents the manifest for a link.d entry.
type LinkManifest struct {
	UserPath         string   `json:"user_path"`                   // absolute path (symlink location)
	Type             LinkType `json:"type"`                        // file or dir
	PackageID        string   `json:"package_id,omitempty"`         // optional package association
	PackageProvider  string   `json:"package_provider,omitempty"` // optional package association
}

// LinkEntry represents a link.d entry (manifest + name).
type LinkEntry struct {
	Name     string        // directory name under link.d (user-given name)
	Manifest *LinkManifest
}

// GetLinkDir returns the path to ~/.al/link.d/
func GetLinkDir() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "link.d"), nil
}

// safeLinkName matches allowed characters for link.d directory name: alphanumeric, underscore, hyphen, dot.
var safeLinkName = regexp.MustCompile(`^[a-zA-Z0-9_.-]+$`)

// sanitizeLinkName validates and returns the name for use as link.d/<name>. Rejects empty and unsafe strings.
func sanitizeLinkName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", fmt.Errorf("link name cannot be empty")
	}
	if strings.Contains(name, "/") || name == "." || name == ".." || strings.HasPrefix(name, "..") {
		return "", fmt.Errorf("invalid link name: %s", name)
	}
	if !safeLinkName.MatchString(name) {
		return "", fmt.Errorf("link name may only contain letters, numbers, underscore, hyphen, and dot: %s", name)
	}
	return name, nil
}

// resolveUserPath returns the absolute path for the user-facing path.
func resolveUserPath(path string) (string, error) {
	if filepath.IsAbs(path) {
		return filepath.Clean(path), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	if path == "~" || strings.HasPrefix(path, "~/") {
		return filepath.Join(home, strings.TrimPrefix(path, "~/")), nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Abs(filepath.Join(cwd, path))
}

// DetectLinkType determines file or dir from path. If path does not exist, trailing / means dir, else file.
func DetectLinkType(userPath string) (LinkType, error) {
	abs, err := resolveUserPath(userPath)
	if err != nil {
		return "", err
	}
	fi, err := os.Stat(abs)
	if err == nil {
		if fi.IsDir() {
			return LinkTypeDir, nil
		}
		return LinkTypeFile, nil
	}
	if !os.IsNotExist(err) {
		return "", err
	}
	// Path does not exist: trailing / => dir, else file
	if strings.HasSuffix(strings.TrimRight(userPath, " "), "/") {
		return LinkTypeDir, nil
	}
	return LinkTypeFile, nil
}

// AddLink adds a new link: copies content to link.d/<name>/content, creates symlink at userPath.
// name is the link name (used as directory under link.d). userPath is the symlink location (can be ~/...). packageID and packageProvider are optional.
func AddLink(name, userPath string, linkType LinkType, packageID, packageProvider string) (*LinkEntry, error) {
	safeName, err := sanitizeLinkName(name)
	if err != nil {
		return nil, err
	}
	absUserPath, err := resolveUserPath(userPath)
	if err != nil {
		return nil, err
	}
	if err := EnsureConfigDir(); err != nil {
		return nil, err
	}
	linkDir, err := GetLinkDir()
	if err != nil {
		return nil, err
	}
	entryDir := filepath.Join(linkDir, safeName)
	if _, err := os.Stat(entryDir); err == nil {
		return nil, fmt.Errorf("link name already exists: %s", safeName)
	}
	contentPath := filepath.Join(entryDir, linkContentName)
	if err := os.MkdirAll(entryDir, 0755); err != nil {
		return nil, err
	}
	manifest := &LinkManifest{
		UserPath:        absUserPath,
		Type:            linkType,
		PackageID:       packageID,
		PackageProvider: packageProvider,
	}
	if linkType == LinkTypeFile {
		// Copy file to link.d/<name>/content
		if _, err := os.Stat(absUserPath); err == nil {
			src, err := os.Open(absUserPath)
			if err != nil {
				return nil, err
			}
			defer src.Close()
			dst, err := os.Create(contentPath)
			if err != nil {
				return nil, err
			}
			if _, err := io.Copy(dst, src); err != nil {
				dst.Close()
				return nil, err
			}
			if err := dst.Close(); err != nil {
				return nil, err
			}
		} else {
			// Create empty file so symlink target exists
			if err := os.WriteFile(contentPath, nil, 0644); err != nil {
				return nil, err
			}
		}
	} else {
		// Dir: content is link.d/<name>/content/ (a directory)
		if _, err := os.Stat(absUserPath); err == nil {
			if err := copyDir(absUserPath, contentPath); err != nil {
				return nil, err
			}
		} else {
			if err := os.MkdirAll(contentPath, 0755); err != nil {
				return nil, err
			}
		}
	}
	if err := saveLinkManifest(entryDir, manifest); err != nil {
		os.RemoveAll(entryDir)
		return nil, err
	}
	// Remove original so we can create symlink (for existing paths)
	if _, err := os.Stat(absUserPath); err == nil {
		if err := os.RemoveAll(absUserPath); err != nil {
			os.RemoveAll(entryDir)
			return nil, fmt.Errorf("removing original for symlink: %w", err)
		}
	} else {
		// Ensure parent dir exists for new path
		if err := os.MkdirAll(filepath.Dir(absUserPath), 0755); err != nil {
			os.RemoveAll(entryDir)
			return nil, err
		}
	}
	if err := os.Symlink(contentPath, absUserPath); err != nil {
		os.RemoveAll(entryDir)
		return nil, fmt.Errorf("creating symlink: %w", err)
	}
	return &LinkEntry{Name: safeName, Manifest: manifest}, nil
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		destPath := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(destPath, info.Mode())
		}
		return copyFile(path, destPath)
	})
}

func copyFile(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Chmod(info.Mode())
}

func loadLinkManifest(entryDir string) (*LinkManifest, error) {
	p := filepath.Join(entryDir, linkManifestFilename)
	data, err := os.ReadFile(p)
	if err != nil {
		return nil, err
	}
	var m LinkManifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

func saveLinkManifest(entryDir string, m *LinkManifest) error {
	p := filepath.Join(entryDir, linkManifestFilename)
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0644)
}

// ListLinks returns all link.d entries. If packageID and packageProvider are both non-empty, filter by that package.
func ListLinks(packageID, packageProvider string) ([]LinkEntry, error) {
	linkDir, err := GetLinkDir()
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(linkDir); os.IsNotExist(err) {
		return nil, nil
	}
	entries, err := os.ReadDir(linkDir)
	if err != nil {
		return nil, err
	}
	var result []LinkEntry
	for _, e := range entries {
		if !e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		entryDir := filepath.Join(linkDir, e.Name())
		m, err := loadLinkManifest(entryDir)
		if err != nil {
			continue
		}
		if packageID != "" && packageProvider != "" {
			if m.PackageID != packageID || m.PackageProvider != packageProvider {
				continue
			}
		}
		result = append(result, LinkEntry{Name: e.Name(), Manifest: m})
	}
	return result, nil
}

// GetLinkByName returns the link entry and its directory by name. Returns nil if not found.
func GetLinkByName(name string) (*LinkEntry, string, error) {
	safeName, err := sanitizeLinkName(name)
	if err != nil {
		return nil, "", err
	}
	linkDir, err := GetLinkDir()
	if err != nil {
		return nil, "", err
	}
	entryDir := filepath.Join(linkDir, safeName)
	if _, err := os.Stat(entryDir); os.IsNotExist(err) {
		return nil, "", nil
	}
	m, err := loadLinkManifest(entryDir)
	if err != nil {
		return nil, "", err
	}
	return &LinkEntry{Name: safeName, Manifest: m}, entryDir, nil
}

// GetLinkContentPath returns the path to the link content (file or dir) inside link.d/<name>/.
func GetLinkContentPath(entryDir string) string {
	return filepath.Join(entryDir, linkContentName)
}

// RemoveLink removes the symlink and optionally copies content back (copy-back). If purge is true, deletes link.d/<name> without copy-back.
func RemoveLink(entry *LinkEntry, entryDir string, purge bool) error {
	contentPath := GetLinkContentPath(entryDir)
	userPath := entry.Manifest.UserPath
	// Remove symlink at user path
	if _, err := os.Lstat(userPath); err == nil {
		if err := os.Remove(userPath); err != nil {
			return fmt.Errorf("removing symlink: %w", err)
		}
	}
	if !purge {
		// Copy content back to user path
		if entry.Manifest.Type == LinkTypeDir {
			if err := os.MkdirAll(userPath, 0755); err != nil {
				return err
			}
			if err := copyDir(contentPath, userPath); err != nil {
				return err
			}
		} else {
			if _, err := os.Stat(contentPath); err == nil {
				if err := os.MkdirAll(filepath.Dir(userPath), 0755); err != nil {
					return err
				}
				if err := copyFile(contentPath, userPath); err != nil {
					return err
				}
			}
		}
	}
	return os.RemoveAll(entryDir)
}

// LinksByPackage returns links that are associated with the given package (id, provider).
func LinksByPackage(packageID, packageProvider string) ([]LinkEntry, error) {
	return ListLinks(packageID, packageProvider)
}

// ClearLinkPackageAssociation clears the package association from a link's manifest (symlink and content are kept).
func ClearLinkPackageAssociation(entryDir string) error {
	m, err := loadLinkManifest(entryDir)
	if err != nil {
		return err
	}
	m.PackageID = ""
	m.PackageProvider = ""
	return saveLinkManifest(entryDir, m)
}
