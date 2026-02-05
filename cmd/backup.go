package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/kkato1030/al/internal/config"
	"github.com/kkato1030/al/internal/source"
	"github.com/spf13/cobra"
)

const defaultBackupRepoName = "dotal"

// NewBackupCmd creates the backup command
func NewBackupCmd() *cobra.Command {
	var init bool
	var repo string

	cmd := &cobra.Command{
		Use:   "backup",
		Short: "Backup al settings to GitHub",
		Long:  "Commit and push ~/.al (AL_HOME) to a GitHub repository. Default repo is owner/dotal (owner from gh). Use --init to create the repository on GitHub first.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBackup(init, repo)
		},
	}

	cmd.Flags().BoolVar(&init, "init", false, "Create the GitHub repository if it does not exist, then push")
	cmd.Flags().StringVar(&repo, "repo", "", "Override backup repository (owner/repo, e.g. kkato1030/dotal)")

	return cmd
}

func runBackup(doInit bool, repoOverride string) error {
	configDir, err := config.GetConfigDir()
	if err != nil {
		return fmt.Errorf("failed to get config directory: %w", err)
	}

	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		return fmt.Errorf("config directory does not exist: %s (run 'al init' first)", configDir)
	}

	owner, repo, err := resolveBackupRepo(repoOverride)
	if err != nil {
		return err
	}

	remoteURL := fmt.Sprintf("https://github.com/%s/%s.git", owner, repo)

	if doInit {
		if err := ensureRepoOnGitHub(owner, repo); err != nil {
			return err
		}
		if err := ensureGitInConfigDir(configDir, remoteURL); err != nil {
			return err
		}
	}

	// Ensure we have git and remote for push
	if err := ensureGitInConfigDir(configDir, remoteURL); err != nil {
		return err
	}

	// Add, commit if there are changes, then push
	if err := gitAddAll(configDir); err != nil {
		return err
	}

	hasChanges, err := gitHasChanges(configDir)
	if err != nil {
		return err
	}
	if hasChanges {
		if err := gitCommit(configDir, "al backup"); err != nil {
			return err
		}
	}

	if err := gitPush(configDir, owner, repo); err != nil {
		return err
	}

	if hasChanges {
		fmt.Printf("Committed and pushed to https://github.com/%s/%s\n", owner, repo)
	} else {
		fmt.Printf("Already up to date at https://github.com/%s/%s\n", owner, repo)
	}
	return nil
}

func resolveBackupRepo(override string) (owner, repo string, err error) {
	if override != "" {
		return source.ParseOwnerRepo(override)
	}
	appCfg, err := config.LoadAppConfig()
	if err != nil {
		return "", "", fmt.Errorf("load config: %w", err)
	}
	if appCfg.BackupRepo != "" {
		return source.ParseOwnerRepo(appCfg.BackupRepo)
	}
	// Default: owner/dotal where owner = gh api user
	login, err := getGitHubLogin()
	if err != nil {
		return "", "", fmt.Errorf("default backup repo is owner/dotal but could not get GitHub username: %w (set backup_repo in config or use --repo)", err)
	}
	return strings.TrimSpace(login), defaultBackupRepoName, nil
}

func getGitHubLogin() (string, error) {
	cmd := exec.Command("gh", "api", "user", "-q", ".login")
	cmd.Stderr = nil
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func ensureRepoOnGitHub(owner, repo string) error {
	// gh repo create owner/repo --public (idempotent: fails if repo exists)
	cmd := exec.Command("gh", "repo", "create", owner+"/"+repo, "--public", "--description", "al config backup")
	cmd.Stdout = nil
	cmd.Stderr = nil
	err := cmd.Run()
	if err != nil {
		// Repo might already exist; check
		checkCmd := exec.Command("gh", "repo", "view", owner+"/"+repo)
		checkCmd.Stdout = nil
		checkCmd.Stderr = nil
		if checkErr := checkCmd.Run(); checkErr == nil {
			return nil
		}
		return fmt.Errorf("create GitHub repo: %w", err)
	}
	return nil
}

func ensureGitInConfigDir(configDir, remoteURL string) error {
	gitDir := filepath.Join(configDir, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		if err := runGit(configDir, "init"); err != nil {
			return fmt.Errorf("git init: %w", err)
		}
		// Default .gitignore to avoid noisy files
		gitignore := filepath.Join(configDir, ".gitignore")
		if _, err := os.Stat(gitignore); os.IsNotExist(err) {
			_ = os.WriteFile(gitignore, []byte(".DS_Store\n"), 0644)
		}
	}

	// Check remote origin
	out, err := runGitOutput(configDir, "remote", "get-url", "origin")
	if err != nil || strings.TrimSpace(string(out)) != remoteURL {
		_ = runGit(configDir, "remote", "remove", "origin")
		if err := runGit(configDir, "remote", "add", "origin", remoteURL); err != nil {
			return fmt.Errorf("git remote add origin: %w", err)
		}
	}
	return nil
}

func gitAddAll(configDir string) error {
	return runGit(configDir, "add", "-A")
}

func gitHasChanges(configDir string) (bool, error) {
	// Check if there are staged changes or untracked files that would be committed
	out, err := runGitOutput(configDir, "status", "--porcelain")
	if err != nil {
		return false, err
	}
	return len(bytes.TrimSpace(out)) > 0, nil
}

func gitCommit(configDir, message string) error {
	return runGit(configDir, "commit", "-m", message)
}

func gitPush(configDir, owner, repo string) error {
	// Ensure branch main exists and push
	if err := runGit(configDir, "branch", "-M", "main"); err != nil {
		return fmt.Errorf("git branch -M main: %w", err)
	}
	if err := runGit(configDir, "push", "-u", "origin", "main"); err != nil {
		return fmt.Errorf("git push: %w", err)
	}
	return nil
}

func runGit(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}

func runGitOutput(dir string, args ...string) ([]byte, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Stderr = nil
	return cmd.Output()
}
