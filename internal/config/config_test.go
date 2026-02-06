package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestGetDefaultAliases(t *testing.T) {
	got := GetDefaultAliases()
	want := map[string]string{
		"promote":  "package move {args} --to package.promote_to",
		"add":      "package add",
		"remove":   "package remove",
		"list":     "package list",
		"register": "package add --provider manual",
		"import":   "package import {args}",
		"pkg":      "package",
		"prf":      "profile",
		"prv":      "provider",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("GetDefaultAliases() = %v, want %v", got, want)
	}
}

func TestGetConfigDir(t *testing.T) {
	dir := t.TempDir()
	restore := setEnv("AL_HOME", dir)
	defer restore()

	got, err := GetConfigDir()
	if err != nil {
		t.Fatal(err)
	}
	if got != dir {
		t.Errorf("GetConfigDir() = %q, want %q", got, dir)
	}
}

func TestGetProvidersConfigPath(t *testing.T) {
	dir := t.TempDir()
	restore := setEnv("AL_HOME", dir)
	defer restore()

	got, err := GetProvidersConfigPath()
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(dir, "providers.json")
	if got != want {
		t.Errorf("GetProvidersConfigPath() = %q, want %q", got, want)
	}
}

func TestGetConfigPath(t *testing.T) {
	dir := t.TempDir()
	restore := setEnv("AL_HOME", dir)
	defer restore()

	got, err := GetConfigPath()
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(dir, "config.json")
	if got != want {
		t.Errorf("GetConfigPath() = %q, want %q", got, want)
	}
}

func TestLoadAppConfig_NoFile(t *testing.T) {
	dir := t.TempDir()
	restore := setEnv("AL_HOME", dir)
	defer restore()

	cfg, err := LoadAppConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg == nil {
		t.Fatal("LoadAppConfig() returned nil")
	}
	if cfg.DefaultProvider != "" || cfg.DefaultProfile != "" {
		t.Errorf("expected empty config, got %+v", cfg)
	}
}

func TestSaveAppConfig_LoadAppConfig_Roundtrip(t *testing.T) {
	dir := t.TempDir()
	restore := setEnv("AL_HOME", dir)
	defer restore()

	want := &AppConfig{
		DefaultProvider: "brew",
		DefaultProfile:  "default",
		DefaultStage:    "stable",
		BackupRepo:      "user/dotal",
	}
	if err := SaveAppConfig(want); err != nil {
		t.Fatal(err)
	}
	got, err := LoadAppConfig()
	if err != nil {
		t.Fatal(err)
	}
	if got.DefaultProvider != want.DefaultProvider ||
		got.DefaultProfile != want.DefaultProfile ||
		got.DefaultStage != want.DefaultStage ||
		got.BackupRepo != want.BackupRepo {
		t.Errorf("LoadAppConfig() = %+v, want %+v", got, want)
	}
}

func setEnv(key, value string) func() {
	old := os.Getenv(key)
	os.Setenv(key, value)
	return func() {
		if old == "" {
			os.Unsetenv(key)
		} else {
			os.Setenv(key, old)
		}
	}
}
