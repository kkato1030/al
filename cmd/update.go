package cmd

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

const (
	githubOwner = "kkato1030"
	githubRepo  = "al"
	githubAPI   = "https://api.github.com"
)

type Release struct {
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
	Assets  []Asset `json:"assets"`
}

type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// NewUpdateCmd creates the update command
func NewUpdateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Check for updates and update al to the latest version",
		Long:  "Check for the latest version of al and update if a newer version is available.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpdate()
		},
	}
}

func runUpdate() error {
	fmt.Println("Checking for updates...")

	// Get current version
	currentVersion := version
	if currentVersion == "dev" {
		return fmt.Errorf("cannot update dev version")
	}

	// Get latest release
	latestRelease, err := getLatestRelease()
	if err != nil {
		return fmt.Errorf("failed to get latest release: %w", err)
	}

	latestVersion := strings.TrimPrefix(latestRelease.TagName, "v")
	currentVersion = strings.TrimPrefix(currentVersion, "v")

	fmt.Printf("Current version: %s\n", currentVersion)
	fmt.Printf("Latest version:  %s\n", latestVersion)

	// Compare versions
	if !isNewerVersion(latestVersion, currentVersion) {
		fmt.Println("You are already using the latest version!")
		return nil
	}

	// Ask for confirmation
	fmt.Printf("\nUpdate available! Do you want to update from %s to %s? [y/N]: ", currentVersion, latestVersion)
	var response string
	fmt.Scanln(&response)
	if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
		fmt.Println("Update cancelled.")
		return nil
	}

	// Perform update
	return performUpdate(latestRelease)
}

func getLatestRelease() (*Release, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/releases/latest", githubAPI, githubOwner, githubRepo)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get latest release: status %d", resp.StatusCode)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	return &release, nil
}

func isNewerVersion(latest, current string) bool {
	// Simple version comparison
	// Remove 'v' prefix if present
	latest = strings.TrimPrefix(latest, "v")
	current = strings.TrimPrefix(current, "v")

	// Split version strings
	latestParts := strings.Split(latest, ".")
	currentParts := strings.Split(current, ".")

	// Compare each part
	maxLen := len(latestParts)
	if len(currentParts) > maxLen {
		maxLen = len(currentParts)
	}

	for i := 0; i < maxLen; i++ {
		var latestPart, currentPart string
		if i < len(latestParts) {
			latestPart = latestParts[i]
		}
		if i < len(currentParts) {
			currentPart = currentParts[i]
		}

		// Compare as strings (works for most cases)
		if latestPart > currentPart {
			return true
		}
		if latestPart < currentPart {
			return false
		}
	}

	return false
}

func performUpdate(release *Release) error {
	// Get current binary path
	currentBinary, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get current binary path: %w", err)
	}

	// Resolve symlink if needed
	currentBinary, err = filepath.EvalSymlinks(currentBinary)
	if err != nil {
		return fmt.Errorf("failed to resolve symlink: %w", err)
	}

	// Determine architecture
	arch := runtime.GOARCH
	if arch == "amd64" {
		// Already correct
	} else if arch == "arm64" {
		// Already correct
	} else {
		return fmt.Errorf("unsupported architecture: %s", arch)
	}

	// Determine OS
	goos := runtime.GOOS
	if goos != "darwin" && goos != "linux" {
		return fmt.Errorf("unsupported OS: %s", goos)
	}

	// Find matching asset
	assetName := fmt.Sprintf("al_%s_%s.tar.gz", goos, arch)
	var downloadURL string
	for _, asset := range release.Assets {
		if asset.Name == assetName {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}

	if downloadURL == "" {
		return fmt.Errorf("no matching asset found for %s/%s", goos, arch)
	}

	fmt.Printf("Downloading %s...\n", assetName)

	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "al-update-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Download the archive
	archivePath := filepath.Join(tmpDir, assetName)
	if err := downloadFile(downloadURL, archivePath); err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}

	// Extract the archive
	binaryPath := filepath.Join(tmpDir, "al")
	if err := extractTarGz(archivePath, tmpDir); err != nil {
		return fmt.Errorf("failed to extract: %w", err)
	}

	// Verify the extracted binary exists
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		return fmt.Errorf("binary not found in archive")
	}

	// Make binary executable
	if err := os.Chmod(binaryPath, 0755); err != nil {
		return fmt.Errorf("failed to make binary executable: %w", err)
	}

	// Check if we need sudo
	needsSudo := false
	if err := os.WriteFile(currentBinary, []byte("test"), 0644); err != nil {
		needsSudo = true
	}

	if needsSudo {
		// Use sudo to replace the binary
		fmt.Println("Installing update (requires sudo)...")
		cmd := exec.Command("sudo", "mv", binaryPath, currentBinary)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to install update: %w", err)
		}
	} else {
		// Replace the binary directly
		fmt.Println("Installing update...")
		if err := os.Rename(binaryPath, currentBinary); err != nil {
			return fmt.Errorf("failed to install update: %w", err)
		}
	}

	fmt.Printf("Successfully updated to version %s!\n", strings.TrimPrefix(release.TagName, "v"))
	return nil
}

func downloadFile(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download: status %d", resp.StatusCode)
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func extractTarGz(archivePath, destDir string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// Only extract the binary file
		if header.Name == "al" {
			target := filepath.Join(destDir, header.Name)
			outFile, err := os.Create(target)
			if err != nil {
				return err
			}

			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		}
	}

	return nil
}
