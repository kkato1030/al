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
	cmd := exec.Command(alBinary, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "AL_HOME="+dir)
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
