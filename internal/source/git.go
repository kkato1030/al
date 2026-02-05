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
