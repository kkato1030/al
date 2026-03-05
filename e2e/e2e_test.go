package e2e

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

var alBinary string

func TestMain(m *testing.M) {
	if os.Getenv("GITHUB_ACTIONS") != "true" {
		os.Exit(0)
	}
	dir, err := os.MkdirTemp("", "al-e2e-build-")
	if err != nil {
		os.Stderr.WriteString("e2e: failed to create temp dir: " + err.Error() + "\n")
		os.Exit(1)
	}
	defer os.RemoveAll(dir)

	binaryPath := filepath.Join(dir, "al")
	wd, err := os.Getwd()
	if err != nil {
		os.Stderr.WriteString("e2e: getwd: " + err.Error() + "\n")
		os.Exit(1)
	}
	if filepath.Base(wd) == "e2e" {
		wd = filepath.Dir(wd)
	}
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Dir = wd
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		os.Stderr.WriteString("e2e: go build failed: " + err.Error() + "\n")
		os.Exit(1)
	}
	alBinary = binaryPath
	os.Exit(m.Run())
}

// runAl runs the al binary with AL_HOME=dir and args, returns stdout, stderr, and exit code.
func runAl(t *testing.T, dir string, args ...string) (stdout, stderr string, code int) {
	t.Helper()
	return runAlWithStdin(t, dir, "", args...)
}

// runAlWithStdin is like runAl but feeds stdin to the command (e.g. for confirmations).
func runAlWithStdin(t *testing.T, dir string, stdin string, args ...string) (stdout, stderr string, code int) {
	t.Helper()
	cmd := exec.Command(alBinary, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "AL_HOME="+dir)
	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err := cmd.Run()
	stdout = outBuf.String()
	stderr = errBuf.String()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			code = exitErr.ExitCode()
		} else {
			t.Fatalf("runAl: %v", err)
		}
	}
	return stdout, stderr, code
}

func TestE2E_Init(t *testing.T) {
	dir := t.TempDir()
	_, _, code := runAl(t, dir, "init")
	if code != 0 {
		t.Fatalf("al init exited with code %d", code)
	}
	for _, name := range []string{"profiles.json", "config.json", ".gitignore", "providers.json"} {
		p := filepath.Join(dir, name)
		if _, err := os.Stat(p); os.IsNotExist(err) {
			t.Errorf("expected %s to exist", name)
		}
	}
}

func TestE2E_InitThenList(t *testing.T) {
	dir := t.TempDir()
	_, _, code := runAl(t, dir, "init")
	if code != 0 {
		t.Fatalf("al init exited with code %d", code)
	}
	_, _, code = runAl(t, dir, "list")
	if code != 0 {
		t.Fatalf("al list exited with code %d", code)
	}
}

func TestE2E_InitThenAddManual(t *testing.T) {
	dir := t.TempDir()
	_, _, code := runAl(t, dir, "init")
	if code != 0 {
		t.Fatalf("al init exited with code %d", code)
	}
	_, _, code = runAl(t, dir, "provider", "add", "manual")
	if code != 0 {
		t.Fatalf("al provider add manual exited with code %d", code)
	}
	_, _, code = runAl(t, dir, "add", "--prv", "manual", "e2e-pkg")
	if code != 0 {
		t.Fatalf("al add --prv manual e2e-pkg exited with code %d", code)
	}
	packagesPath := filepath.Join(dir, "packages.json")
	data, err := os.ReadFile(packagesPath)
	if err != nil {
		t.Fatalf("read packages.json: %v", err)
	}
	var cfg struct {
		Packages []struct {
			Name     string `json:"name"`
			Provider string `json:"provider"`
		} `json:"packages"`
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("parse packages.json: %v", err)
	}
	var found bool
	for _, p := range cfg.Packages {
		if p.Name == "e2e-pkg" && p.Provider == "manual" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("packages.json: expected entry for e2e-pkg (manual), got %+v", cfg.Packages)
	}
}

func TestE2E_InitThenConfigShow(t *testing.T) {
	dir := t.TempDir()
	_, _, code := runAl(t, dir, "init")
	if code != 0 {
		t.Fatalf("al init exited with code %d", code)
	}
	stdout, _, code := runAl(t, dir, "config", "show")
	if code != 0 {
		t.Fatalf("al config show exited with code %d", code)
	}
	if !strings.Contains(stdout, "core.trial") && !strings.Contains(stdout, "default_profile") {
		t.Errorf("al config show output should contain default info; got: %s", stdout)
	}
}

func TestE2E_InitThenDoctor(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("e2e doctor requires darwin")
	}
	dir := t.TempDir()
	_, _, code := runAl(t, dir, "init")
	if code != 0 {
		t.Fatalf("al init exited with code %d", code)
	}
	_, stderr, code := runAl(t, dir, "doctor")
	if code != 0 {
		t.Logf("al doctor stderr: %s", stderr)
		t.Fatalf("al doctor exited with code %d", code)
	}
}

// --- Section 1: Easy additions (non-interactive, no external deps) ---

func TestE2E_Version(t *testing.T) {
	dir := t.TempDir()
	stdout, _, code := runAl(t, dir, "version")
	if code != 0 {
		t.Fatalf("al version exited with code %d", code)
	}
	if strings.TrimSpace(stdout) == "" {
		t.Error("al version should output something")
	}
}

func TestE2E_ConfigAliasList(t *testing.T) {
	dir := t.TempDir()
	_, _, code := runAl(t, dir, "init")
	if code != 0 {
		t.Fatalf("al init exited with code %d", code)
	}
	stdout, _, code := runAl(t, dir, "config", "alias", "list")
	if code != 0 {
		t.Fatalf("al config alias list exited with code %d", code)
	}
	if !strings.Contains(stdout, "add") && !strings.Contains(stdout, "package") {
		t.Errorf("alias list should show aliases; got: %s", stdout)
	}
}

func TestE2E_ProviderList(t *testing.T) {
	dir := t.TempDir()
	_, _, code := runAl(t, dir, "init")
	if code != 0 {
		t.Fatalf("al init exited with code %d", code)
	}
	stdout, _, code := runAl(t, dir, "provider", "list")
	if code != 0 {
		t.Fatalf("al provider list exited with code %d", code)
	}
	if !strings.Contains(stdout, "brew") {
		t.Errorf("provider list should contain brew; got: %s", stdout)
	}
}

func TestE2E_ProfileAddListShow(t *testing.T) {
	dir := t.TempDir()
	_, _, code := runAl(t, dir, "init")
	if code != 0 {
		t.Fatalf("al init exited with code %d", code)
	}
	_, _, code = runAl(t, dir, "profile", "add", "work", "-d", "e2e profile")
	if code != 0 {
		t.Fatalf("al profile add work exited with code %d", code)
	}
	stdout, _, code := runAl(t, dir, "profile", "list")
	if code != 0 {
		t.Fatalf("al profile list exited with code %d", code)
	}
	if !strings.Contains(stdout, "work") {
		t.Errorf("profile list should contain work; got: %s", stdout)
	}
	stdout, _, code = runAl(t, dir, "profile", "show", "work")
	if code != 0 {
		t.Fatalf("al profile show work exited with code %d", code)
	}
	if !strings.Contains(stdout, "work") {
		t.Errorf("profile show work should mention work; got: %s", stdout)
	}
}

func TestE2E_PackageShow(t *testing.T) {
	dir := t.TempDir()
	_, _, code := runAl(t, dir, "init")
	if code != 0 {
		t.Fatalf("al init exited with code %d", code)
	}
	_, _, code = runAl(t, dir, "provider", "add", "manual")
	if code != 0 {
		t.Fatalf("al provider add manual exited with code %d", code)
	}
	_, _, code = runAl(t, dir, "add", "--prv", "manual", "e2e-pkg")
	if code != 0 {
		t.Fatalf("al add manual e2e-pkg exited with code %d", code)
	}
	stdout, _, code := runAl(t, dir, "package", "show", "e2e-pkg")
	if code != 0 {
		t.Fatalf("al package show e2e-pkg exited with code %d", code)
	}
	if !strings.Contains(stdout, "e2e-pkg") || !strings.Contains(stdout, "manual") {
		t.Errorf("package show should show e2e-pkg and manual; got: %s", stdout)
	}
}

func TestE2E_PackageRemove(t *testing.T) {
	dir := t.TempDir()
	_, _, code := runAl(t, dir, "init")
	if code != 0 {
		t.Fatalf("al init exited with code %d", code)
	}
	_, _, code = runAl(t, dir, "provider", "add", "manual")
	if code != 0 {
		t.Fatalf("al provider add manual exited with code %d", code)
	}
	_, _, code = runAl(t, dir, "add", "--prv", "manual", "e2e-pkg")
	if code != 0 {
		t.Fatalf("al add manual e2e-pkg exited with code %d", code)
	}
	// manual provider remove prompts "Have you already uninstalled? [y/N]"; feed "y"
	_, _, code = runAlWithStdin(t, dir, "y\n", "remove", "e2e-pkg", "--prv", "manual", "--prf", "core")
	if code != 0 {
		t.Fatalf("al remove e2e-pkg exited with code %d", code)
	}
	data, err := os.ReadFile(filepath.Join(dir, "packages.json"))
	if err != nil {
		t.Fatalf("read packages.json: %v", err)
	}
	var cfg struct {
		Packages []struct {
			Name string `json:"name"`
		} `json:"packages"`
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("parse packages.json: %v", err)
	}
	for _, p := range cfg.Packages {
		if p.Name == "e2e-pkg" {
			t.Error("e2e-pkg should be removed from packages.json")
		}
	}
}

func TestE2E_ConfigSet(t *testing.T) {
	dir := t.TempDir()
	_, _, code := runAl(t, dir, "init")
	if code != 0 {
		t.Fatalf("al init exited with code %d", code)
	}
	_, _, code = runAl(t, dir, "config", "set", "--default-profile", "core.trial")
	if code != 0 {
		t.Fatalf("al config set exited with code %d", code)
	}
	stdout, _, code := runAl(t, dir, "config", "show")
	if code != 0 {
		t.Fatalf("al config show exited with code %d", code)
	}
	if !strings.Contains(stdout, "core.trial") {
		t.Errorf("config show should contain core.trial after set; got: %s", stdout)
	}
}

func TestE2E_ActivateZsh(t *testing.T) {
	dir := t.TempDir()
	_, _, code := runAl(t, dir, "init")
	if code != 0 {
		t.Fatalf("al init exited with code %d", code)
	}
	stdout, _, code := runAl(t, dir, "activate", "zsh")
	if code != 0 {
		t.Fatalf("al activate zsh exited with code %d", code)
	}
	if strings.TrimSpace(stdout) == "" {
		t.Error("al activate zsh should output shell code")
	}
}

func TestE2E_LogsList(t *testing.T) {
	dir := t.TempDir()
	_, _, code := runAl(t, dir, "init")
	if code != 0 {
		t.Fatalf("al init exited with code %d", code)
	}
	_, _, code = runAl(t, dir, "logs", "--list")
	if code != 0 {
		t.Fatalf("al logs --list exited with code %d", code)
	}
}

func TestE2E_ProfileTemplateList(t *testing.T) {
	dir := t.TempDir()
	_, _, code := runAl(t, dir, "init")
	if code != 0 {
		t.Fatalf("al init exited with code %d", code)
	}
	_, _, code = runAl(t, dir, "profile", "template", "list")
	if code != 0 {
		t.Fatalf("al profile template list exited with code %d", code)
	}
}

// --- Section 2: sync / backup dry-run (no GitHub) ---

func TestE2E_SyncPlan(t *testing.T) {
	dir := t.TempDir()
	_, _, code := runAl(t, dir, "init")
	if code != 0 {
		t.Fatalf("al init exited with code %d", code)
	}
	stdout, _, code := runAl(t, dir, "sync", "--plan")
	if code != 0 {
		t.Fatalf("al sync --plan exited with code %d", code)
	}
	if !strings.Contains(stdout, "Plan") && !strings.Contains(stdout, "plan") && !strings.Contains(stdout, "profiles") {
		t.Errorf("sync --plan should output plan; got: %s", stdout)
	}
}

func TestE2E_BackupDryRun(t *testing.T) {
	dir := t.TempDir()
	_, _, code := runAl(t, dir, "init")
	if code != 0 {
		t.Fatalf("al init exited with code %d", code)
	}
	stdout, _, code := runAl(t, dir, "backup", "--dry-run", "--repo", "foo/bar")
	if code != 0 {
		t.Fatalf("al backup --dry-run --repo foo/bar exited with code %d", code)
	}
	if !strings.Contains(stdout, "dry-run") && !strings.Contains(stdout, "foo/bar") {
		t.Errorf("backup --dry-run should show preview; got: %s", stdout)
	}
}

// --- Section 3: link.d ---

func TestE2E_LinkAddList(t *testing.T) {
	dir := t.TempDir()
	linkTarget := filepath.Join(dir, "linksrc")
	if err := os.MkdirAll(linkTarget, 0755); err != nil {
		t.Fatalf("create link target dir: %v", err)
	}
	_, _, code := runAl(t, dir, "init")
	if code != 0 {
		t.Fatalf("al init exited with code %d", code)
	}
	_, _, code = runAl(t, dir, "link", "add", "e2elink", "--path", linkTarget)
	if code != 0 {
		t.Fatalf("al link add e2elink exited with code %d", code)
	}
	stdout, _, code := runAl(t, dir, "link", "list")
	if code != 0 {
		t.Fatalf("al link list exited with code %d", code)
	}
	if !strings.Contains(stdout, "e2elink") {
		t.Errorf("link list should contain e2elink; got: %s", stdout)
	}
}

func TestE2E_SyncLinkOnly(t *testing.T) {
	dir := t.TempDir()
	linkSrc := filepath.Join(dir, "linksrc.txt")
	if err := os.WriteFile(linkSrc, []byte("hello"), 0644); err != nil {
		t.Fatalf("create link source file: %v", err)
	}
	_, _, code := runAl(t, dir, "init")
	if code != 0 {
		t.Fatalf("al init exited with code %d", code)
	}
	// al link add moves the file to link.d/<name>/content and creates a symlink at linkSrc
	_, _, code = runAl(t, dir, "link", "add", "synclink", "--path", linkSrc)
	if code != 0 {
		t.Fatalf("al link add synclink exited with code %d", code)
	}
	// Remove the symlink at linkSrc to simulate restoring on a new machine
	// (link.d/synclink/content still holds the file; only the symlink is gone)
	if err := os.Remove(linkSrc); err != nil {
		t.Fatalf("remove symlink at linkSrc: %v", err)
	}
	// Run sync --link-only: should recreate the symlink at linkSrc
	stdout, _, code := runAl(t, dir, "sync", "--link-only")
	if code != 0 {
		t.Fatalf("al sync --link-only exited with code %d; stdout: %s", code, stdout)
	}
	// Verify symlink was recreated at linkSrc
	fi, err := os.Lstat(linkSrc)
	if err != nil {
		t.Fatalf("expected symlink at %s after sync --link-only, got error: %v", linkSrc, err)
	}
	if fi.Mode()&os.ModeSymlink == 0 {
		t.Errorf("expected %s to be a symlink after sync --link-only, got mode %v", linkSrc, fi.Mode())
	}
	// Verify output includes the link name
	if !strings.Contains(stdout, "synclink") {
		t.Errorf("sync --link-only output should mention link name 'synclink'; got: %s", stdout)
	}
}

func TestE2E_SyncLinkOnlyPlan(t *testing.T) {
	dir := t.TempDir()
	linkSrc := filepath.Join(dir, "planlink.txt")
	if err := os.WriteFile(linkSrc, []byte("data"), 0644); err != nil {
		t.Fatalf("create link source file: %v", err)
	}
	_, _, code := runAl(t, dir, "init")
	if code != 0 {
		t.Fatalf("al init exited with code %d", code)
	}
	// al link add moves the file to link.d/<name>/content and creates a symlink at linkSrc
	_, _, code = runAl(t, dir, "link", "add", "planlink", "--path", linkSrc)
	if code != 0 {
		t.Fatalf("al link add planlink exited with code %d", code)
	}
	stdout, _, code := runAl(t, dir, "sync", "--link-only", "--plan")
	if code != 0 {
		t.Fatalf("al sync --link-only --plan exited with code %d; stdout: %s", code, stdout)
	}
	// Plan output should show the "Links" section with the link's user path
	if !strings.Contains(stdout, "Links") {
		t.Errorf("sync --link-only --plan should show 'Links' section; got: %s", stdout)
	}
	if !strings.Contains(stdout, linkSrc) {
		t.Errorf("sync --link-only --plan should show link path %s; got: %s", linkSrc, stdout)
	}
	// Plan output should NOT mention bootstrap (--link-only skips it)
	if strings.Contains(strings.ToLower(stdout), "bootstrap") {
		t.Errorf("sync --link-only --plan should not mention bootstrap; got: %s", stdout)
	}
	// Plan output should NOT show "Sync target profiles" (no package sync with --link-only)
	if strings.Contains(stdout, "Sync target profiles") {
		t.Errorf("sync --link-only --plan should not show 'Sync target profiles'; got: %s", stdout)
	}
}

// --- Section 4: darwin only (diff) ---

func TestE2E_Diff(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("e2e diff requires darwin")
	}
	dir := t.TempDir()
	_, _, code := runAl(t, dir, "init")
	if code != 0 {
		t.Fatalf("al init exited with code %d", code)
	}
	_, _, code = runAl(t, dir, "provider", "add", "manual")
	if code != 0 {
		t.Fatalf("al provider add manual exited with code %d", code)
	}
	_, _, code = runAl(t, dir, "add", "--prv", "manual", "e2e-pkg")
	if code != 0 {
		t.Fatalf("al add manual e2e-pkg exited with code %d", code)
	}
	_, _, code = runAl(t, dir, "diff")
	// diff exits 0 when no drift, 1 when there is drift
	if code != 0 && code != 1 {
		t.Fatalf("al diff should exit 0 or 1, got %d", code)
	}
}
