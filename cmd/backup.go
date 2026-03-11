package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/kkato1030/al/internal/config"
	"github.com/kkato1030/al/internal/output"
	"github.com/kkato1030/al/internal/source"
	"github.com/spf13/cobra"
)

const defaultBackupRepoName = "dotal"

// NewBackupCmd creates the backup command
func NewBackupCmd() *cobra.Command {
	var init bool
	var private bool
	var repo string
	var dryRun bool
	var pull bool

	cmd := &cobra.Command{
		Use:   "backup",
		Short: "Backup al settings to GitHub",
		Long:  "Commit and push ~/.al (AL_HOME) to a GitHub repository. Default repo is owner/dotal (owner from gh). Use --init to create the repository on GitHub first. Use --dry-run to preview what would be backed up without actually committing or pushing. Use --pull to fetch and merge changes from the remote backup into ~/.al.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if pull {
				return runBackupPull(repo, dryRun)
			}
			return runBackup(init, private, repo, dryRun)
		},
	}

	cmd.Flags().BoolVar(&init, "init", false, "Create the GitHub repository if it does not exist, then push")
	cmd.Flags().BoolVar(&private, "private", false, "With --init, create the repository as private")
	cmd.Flags().StringVar(&repo, "repo", "", "Override backup repository (owner/repo, e.g. kkato1030/dotal)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be backed up without committing or pushing")
	cmd.Flags().BoolVar(&pull, "pull", false, "Fetch and merge changes from the remote backup into ~/.al")

	return cmd
}

func runBackup(doInit bool, private bool, repoOverride string, dryRun bool) error {
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

	if dryRun {
		fmt.Println("[dry-run] Backup preview:")
		fmt.Printf("  Target repository: https://github.com/%s/%s\n", owner, repo)

		if doInit {
			fmt.Printf("  Would create repository (visibility: %s)\n", map[bool]string{true: "private", false: "public"}[private])
			fmt.Println("  Would initialize git in config directory")
		}

		// Check what files would be backed up
		gitDir := filepath.Join(configDir, ".git")
		if _, err := os.Stat(gitDir); os.IsNotExist(err) {
			fmt.Println("  Would initialize git repository")
		}

		// Show what would be added
		fmt.Println("\n  Files that would be backed up:")
		if err := showFilesToBackup(configDir); err != nil {
			return fmt.Errorf("failed to check files: %w", err)
		}

		// Check if there would be changes to commit
		hasChanges, err := checkWouldHaveChanges(configDir)
		if err != nil {
			return fmt.Errorf("failed to check changes: %w", err)
		}

		if hasChanges {
			fmt.Println("\n  Would commit changes with message: \"al backup\"")
			fmt.Printf("  Would push to https://github.com/%s/%s\n", owner, repo)
		} else {
			fmt.Println("\n  No changes to commit (already up to date)")
		}

		return nil
	}

	if doInit {
		if err := ensureRepoOnGitHub(owner, repo, private); err != nil {
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

func ensureRepoOnGitHub(owner, repo string, private bool) error {
	// gh repo create owner/repo (idempotent: fails if repo exists)
	vis := "--public"
	if private {
		vis = "--private"
	}
	cmd := exec.Command("gh", "repo", "create", owner+"/"+repo, vis, "--description", "al config backup")
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

// showFilesToBackup lists files that would be backed up
func showFilesToBackup(configDir string) error {
	gitDir := filepath.Join(configDir, ".git")
	gitInitialized := false
	if _, err := os.Stat(gitDir); err == nil {
		gitInitialized = true
	}

	if !gitInitialized {
		// If git is not initialized, show all files
		entries, err := os.ReadDir(configDir)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			if entry.Name() == ".git" {
				continue
			}
			fmt.Printf("    %s\n", entry.Name())
		}
		return nil
	}

	// If git is initialized, use git status to show what would be added
	out, err := runGitOutput(configDir, "status", "--porcelain")
	if err != nil {
		// If git status fails, fallback to listing files
		entries, err := os.ReadDir(configDir)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			if entry.Name() == ".git" {
				continue
			}
			fmt.Printf("    %s\n", entry.Name())
		}
		return nil
	}

	lines := strings.Split(string(out), "\n")
	hasFiles := false
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		hasFiles = true
		fmt.Printf("    %s\n", line)
	}

	if !hasFiles {
		fmt.Println("    (no changes)")
	}

	return nil
}

// checkWouldHaveChanges checks if there would be changes to commit
func checkWouldHaveChanges(configDir string) (bool, error) {
	gitDir := filepath.Join(configDir, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		// If git is not initialized, there would be changes
		return true, nil
	}

	// Check git status
	out, err := runGitOutput(configDir, "status", "--porcelain")
	if err != nil {
		// If we can't check, assume there are changes
		return true, nil
	}

	return len(bytes.TrimSpace(out)) > 0, nil
}

// runBackupPull fetches and merges changes from the remote backup repository into ~/.al.
// If merge conflicts occur, they are reported and the user is guided to resolve them manually.
func runBackupPull(repoOverride string, dryRun bool) error {
	configDir, err := config.GetConfigDir()
	if err != nil {
		return fmt.Errorf("failed to get config directory: %w", err)
	}

	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		return fmt.Errorf("config directory does not exist: %s (run 'al init' first)", configDir)
	}

	// Ensure the config dir is a git repository
	gitDir := filepath.Join(configDir, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return fmt.Errorf("config directory is not a git repository: %s (run 'al backup --init' first)", configDir)
	}

	owner, repo, err := resolveBackupRepo(repoOverride)
	if err != nil {
		return err
	}

	remoteURL := fmt.Sprintf("https://github.com/%s/%s.git", owner, repo)

	if dryRun {
		fmt.Println("[dry-run] Backup pull preview:")
		fmt.Printf("  Source repository: https://github.com/%s/%s\n", owner, repo)
		fmt.Println("  Would fetch and merge changes from remote into", configDir)

		// Check for local uncommitted changes
		hasLocal, err := checkWouldHaveChanges(configDir)
		if err == nil && hasLocal {
			output.Warning("There are local uncommitted changes -- a pull may cause conflicts")
		}
		return nil
	}

	// Ensure the remote is set correctly
	if err := ensureGitInConfigDir(configDir, remoteURL); err != nil {
		return err
	}

	// Run git pull (fetch + merge)
	pullOut, pullErr := runGitCombinedOutput(configDir, "pull", "origin", "main")
	pullOutput := strings.TrimSpace(string(pullOut))

	if pullErr != nil {
		// Check if the error is due to merge conflicts
		conflictFiles, conflictErr := getConflictedFiles(configDir)
		if conflictErr == nil && len(conflictFiles) > 0 {
			output.Warning("Merge conflicts detected. Resolve the following files and then run 'git add' and 'git commit':")
			for _, f := range conflictFiles {
				fmt.Fprintf(os.Stderr, "  %s\n", f)
			}
			fmt.Fprintf(os.Stderr, "To abort the merge, run: cd %s && git merge --abort\n", configDir)
			return fmt.Errorf("pull failed due to merge conflicts in %s", configDir)
		}
		// Some other pull error
		if pullOutput != "" {
			return fmt.Errorf("git pull failed: %w\n%s", pullErr, pullOutput)
		}
		return fmt.Errorf("git pull failed: %w", pullErr)
	}

	if pullOutput == "" || pullOutput == "Already up to date." {
		fmt.Printf("Already up to date with https://github.com/%s/%s\n", owner, repo)
	} else {
		fmt.Printf("Pulled from https://github.com/%s/%s\n", owner, repo)
		fmt.Println(pullOutput)
	}
	return nil
}

// runGitCombinedOutput runs a git command and returns combined stdout+stderr output.
func runGitCombinedOutput(dir string, args ...string) ([]byte, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	return cmd.CombinedOutput()
}

// getConflictedFiles returns a list of files with merge conflicts.
func getConflictedFiles(configDir string) ([]string, error) {
	out, err := runGitOutput(configDir, "diff", "--name-only", "--diff-filter=U")
	if err != nil {
		return nil, err
	}
	var files []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			files = append(files, line)
		}
	}
	return files, nil
}
