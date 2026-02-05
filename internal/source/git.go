package source

import (
	"fmt"
	"os/exec"
	"strings"
)

// Clone clones a GitHub repository into dest.
// owner and repo are the GitHub owner and repository name (e.g. "kkato1030", "dotfiles").
// dest is the destination directory path (e.g. AL_HOME).
func Clone(dest, owner, repo string) error {
	owner = strings.TrimSpace(owner)
	repo = strings.TrimSpace(repo)
	if owner == "" || repo == "" {
		return fmt.Errorf("owner and repo must be non-empty")
	}
	url := fmt.Sprintf("https://github.com/%s/%s.git", owner, repo)
	cmd := exec.Command("git", "clone", url, dest)
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git clone %s: %w", url, err)
	}
	return nil
}

// ParseOwnerRepo parses "owner/repo" string into owner and repo.
// Returns an error if the format is invalid (e.g. "kkato1030/dotal").
func ParseOwnerRepo(s string) (owner, repo string, err error) {
	s = strings.TrimSpace(s)
	parts := strings.SplitN(s, "/", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("expected owner/repo (e.g. kkato1030/dotfiles), got: %s", s)
	}
	owner = strings.TrimSpace(parts[0])
	repo = strings.TrimSpace(parts[1])
	if owner == "" || repo == "" {
		return "", "", fmt.Errorf("owner and repo must be non-empty, got: %s", s)
	}
	return owner, repo, nil
}
