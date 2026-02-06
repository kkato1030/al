package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadBrewTapsConfig_NoFile(t *testing.T) {
	dir := t.TempDir()
	restore := setEnv("AL_HOME", dir)
	defer restore()

	cfg, err := LoadBrewTapsConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg == nil || len(cfg.Taps) != 0 {
		t.Errorf("LoadBrewTapsConfig() = %v, want empty taps", cfg)
	}
}

func TestAddBrewTap_Save_Load(t *testing.T) {
	dir := t.TempDir()
	restore := setEnv("AL_HOME", dir)
	defer restore()

	if err := EnsureConfigDir(); err != nil {
		t.Fatal(err)
	}

	if err := AddBrewTap("homebrew/cask-fonts"); err != nil {
		t.Fatal(err)
	}
	if err := AddBrewTap("homebrew/cask"); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadBrewTapsConfig()
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Taps) != 2 {
		t.Errorf("LoadBrewTapsConfig() has %d taps, want 2", len(cfg.Taps))
	}

	has, err := HasBrewTap("homebrew/cask-fonts")
	if err != nil || !has {
		t.Errorf("HasBrewTap(homebrew/cask-fonts) = %v, %v; want true, nil", has, err)
	}
}

func TestAddBrewTap_Idempotent(t *testing.T) {
	dir := t.TempDir()
	restore := setEnv("AL_HOME", dir)
	defer restore()

	if err := EnsureConfigDir(); err != nil {
		t.Fatal(err)
	}

	if err := AddBrewTap("user/repo"); err != nil {
		t.Fatal(err)
	}
	if err := AddBrewTap("user/repo"); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadBrewTapsConfig()
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Taps) != 1 || cfg.Taps[0] != "user/repo" {
		t.Errorf("LoadBrewTapsConfig() = %v, want [user/repo]", cfg.Taps)
	}
}

func TestRemoveBrewTap(t *testing.T) {
	dir := t.TempDir()
	restore := setEnv("AL_HOME", dir)
	defer restore()

	if err := EnsureConfigDir(); err != nil {
		t.Fatal(err)
	}
	path, _ := GetBrewTapsConfigPath()
	if err := os.WriteFile(path, []byte(`{"taps":["a/b","c/d"]}`), 0644); err != nil {
		t.Fatal(err)
	}

	if err := RemoveBrewTap("a/b"); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadBrewTapsConfig()
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Taps) != 1 || cfg.Taps[0] != "c/d" {
		t.Errorf("after RemoveBrewTap(a/b), LoadBrewTapsConfig() = %v", cfg.Taps)
	}

	has, err := HasBrewTap("a/b")
	if err != nil || has {
		t.Errorf("HasBrewTap(a/b) = %v, %v; want false, nil", has, err)
	}
}

func TestGetBrewTapsConfigPath(t *testing.T) {
	dir := t.TempDir()
	restore := setEnv("AL_HOME", dir)
	defer restore()

	path, err := GetBrewTapsConfigPath()
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(dir, "brew-taps.json")
	if path != want {
		t.Errorf("GetBrewTapsConfigPath() = %q, want %q", path, want)
	}
}
