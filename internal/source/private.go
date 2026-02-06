package source

import (
	"fmt"
	"os"
	"os/exec"
)

// EnsurePrivateCloneSetup ensures git, gh, and git-credential-manager are installed
// and runs gh auth login so that private GitHub repositories can be cloned.
// See: https://docs.github.com/ja/get-started/git-basics/caching-your-github-credentials-in-git
func EnsurePrivateCloneSetup() error {
	// Ensure brew is available for installing missing tools
	if _, err := exec.LookPath("brew"); err != nil {
		return fmt.Errorf("brew is required for --private; please install Homebrew first")
	}

	tools := []struct {
		name    string
		install string // brew install args
	}{
		{"git", "git"},
		{"gh", "gh"},
		{"git-credential-manager", "git-credential-manager"},
	}

	for _, t := range tools {
		if _, err := exec.LookPath(t.name); err != nil {
			fmt.Printf("Installing %s...\n", t.name)
			var cmd *exec.Cmd
			if t.name == "git-credential-manager" {
				cmd = exec.Command("brew", "install", "--cask", t.install)
			} else {
				cmd = exec.Command("brew", "install", t.install)
			}
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("install %s: %w", t.name, err)
			}
		}
	}

	// Run gh auth login so that private repos are accessible (git will use gh as credential helper)
	fmt.Println("Authenticating with GitHub (gh auth login)...")
	cmd := exec.Command("gh", "auth", "login", "--hostname", "github.com")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("gh auth login: %w", err)
	}
	return nil
}
