package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestValidateProfileName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"empty", "", true},
		{"simple", "default", false},
		{"with stage", "work.stable", false},
		{"with hyphen", "my-profile", false},
		{"with underscore", "my_profile", false},
		{"with hash", "dev#1", false},
		{"with at", "user@host", false},
		{"with dot", "a.b", false},
		{"invalid space", "has space", true},
		{"invalid slash", "has/slash", true},
		{"invalid colon", "has:colon", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateProfileName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateProfileName(%q) err = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestParseProfileName(t *testing.T) {
	tests := []struct {
		name        string
		fullName    string
		wantProfile string
		wantStage   string
		wantErr     bool
	}{
		{"no stage", "default", "default", "", false},
		{"with stage", "work.stable", "work", "stable", false},
		{"empty", "", "", "", true},
		{"invalid char", "has space", "", "", true},
		{"dot only stage", "a.b", "a", "b", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profile, stage, err := ParseProfileName(tt.fullName)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseProfileName() err = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}
			if profile != tt.wantProfile || stage != tt.wantStage {
				t.Errorf("ParseProfileName() = %q, %q, want %q, %q", profile, stage, tt.wantProfile, tt.wantStage)
			}
		})
	}
}

func TestBuildProfileName(t *testing.T) {
	tests := []struct {
		name    string
		profile string
		stage   string
		want    string
		wantErr bool
	}{
		{"no stage", "default", "", "default", false},
		{"with stage", "work", "stable", "work.stable", false},
		{"invalid profile", "", "stable", "", true},
		{"invalid stage", "work", "bad stage", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BuildProfileName(tt.profile, tt.stage)
			if (err != nil) != tt.wantErr {
				t.Fatalf("BuildProfileName() err = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}
			if got != tt.want {
				t.Errorf("BuildProfileName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIsAutoSyncEnabled(t *testing.T) {
	trueVal := true
	falseVal := false
	tests := []struct {
		name string
		p    ProfileConfig
		want bool
	}{
		{"nil AutoSync", ProfileConfig{Name: "a", AutoSync: nil}, true},
		{"true AutoSync", ProfileConfig{Name: "a", AutoSync: &trueVal}, true},
		{"false AutoSync", ProfileConfig{Name: "a", AutoSync: &falseVal}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsAutoSyncEnabled(tt.p); got != tt.want {
				t.Errorf("IsAutoSyncEnabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProfilesConfig_Roundtrip(t *testing.T) {
	dir := t.TempDir()
	restore := setEnv("AL_HOME", dir)
	defer restore()

	trueVal := true
	cfg := &ProfilesConfig{
		Profiles: []ProfileConfig{
			{Name: "default", Stage: "stable"},
			{Name: "work.trial", Extends: []string{"default"}, AutoSync: &trueVal},
		},
	}
	if err := SaveProfilesConfig(cfg); err != nil {
		t.Fatal(err)
	}
	got, err := LoadProfilesConfig()
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Profiles) != len(cfg.Profiles) {
		t.Fatalf("profiles: got %d, want %d", len(got.Profiles), len(cfg.Profiles))
	}
	for i := range cfg.Profiles {
		if got.Profiles[i].Name != cfg.Profiles[i].Name {
			t.Errorf("profile[%d].Name = %q, want %q", i, got.Profiles[i].Name, cfg.Profiles[i].Name)
		}
	}
}

func TestAddOrUpdateProfile_GetProfile_RemoveProfile(t *testing.T) {
	dir := t.TempDir()
	restore := setEnv("AL_HOME", dir)
	defer restore()

	// Ensure config dir exists
	if err := EnsureConfigDir(); err != nil {
		t.Fatal(err)
	}

	// Add profile
	p := ProfileConfig{Name: "test", Description: "desc"}
	if err := AddOrUpdateProfile(p); err != nil {
		t.Fatal(err)
	}
	got, err := GetProfile("test")
	if err != nil || got == nil {
		t.Fatalf("GetProfile: err=%v got=%v", err, got)
	}
	if got.Description != "desc" {
		t.Errorf("GetProfile: Description = %q", got.Description)
	}

	// Update same profile
	p.Description = "updated"
	if err := AddOrUpdateProfile(p); err != nil {
		t.Fatal(err)
	}
	got, _ = GetProfile("test")
	if got.Description != "updated" {
		t.Errorf("after update: Description = %q", got.Description)
	}

	// Remove profile
	if err := RemoveProfile("test"); err != nil {
		t.Fatal(err)
	}
	got, _ = GetProfile("test")
	if got != nil {
		t.Error("GetProfile after RemoveProfile should return nil")
	}
}

func TestResolveProfileWithExtends(t *testing.T) {
	dir := t.TempDir()
	restore := setEnv("AL_HOME", dir)
	defer restore()

	// Write profiles with extends
	path := filepath.Join(dir, "profiles.json")
	content := `{"profiles":[
		{"name":"base"},
		{"name":"mid","extends":["base"]},
		{"name":"top","extends":["mid"]}
	]}`
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	names, err := ResolveProfileWithExtends("top")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"top", "mid", "base"}
	if !reflect.DeepEqual(names, want) {
		t.Errorf("ResolveProfileWithExtends(\"top\") = %v, want %v", names, want)
	}

	// Not found
	_, err = ResolveProfileWithExtends("nonexistent")
	if err == nil {
		t.Error("ResolveProfileWithExtends(\"nonexistent\") expected error")
	}
}

func TestGetSyncTargetProfileNames(t *testing.T) {
	dir := t.TempDir()
	restore := setEnv("AL_HOME", dir)
	defer restore()

	path := filepath.Join(dir, "profiles.json")
	content := `{"profiles":[
		{"name":"a","auto_sync":true},
		{"name":"b","auto_sync":false},
		{"name":"c"}
	]}`
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	t.Run("mode all", func(t *testing.T) {
		names, err := GetSyncTargetProfileNames("all", "")
		if err != nil {
			t.Fatal(err)
		}
		// a (true), c (nil = true); b (false) excluded
		if len(names) != 2 {
			t.Errorf("got %v", names)
		}
	})

	t.Run("mode profile empty name", func(t *testing.T) {
		_, err := GetSyncTargetProfileNames("profile", "")
		if err == nil {
			t.Error("expected error for empty profile name")
		}
	})

	t.Run("mode profile", func(t *testing.T) {
		names, err := GetSyncTargetProfileNames("profile", "a")
		if err != nil {
			t.Fatal(err)
		}
		if len(names) != 1 || names[0] != "a" {
			t.Errorf("got %v", names)
		}
	})

	t.Run("mode default no default_profile", func(t *testing.T) {
		// config.json without default_profile
		cfgPath := filepath.Join(dir, "config.json")
		if err := os.WriteFile(cfgPath, []byte("{}"), 0644); err != nil {
			t.Fatal(err)
		}
		names, err := GetSyncTargetProfileNames("default", "")
		if err != nil {
			t.Fatal(err)
		}
		if len(names) != 0 {
			t.Errorf("got %v", names)
		}
	})

	t.Run("invalid mode", func(t *testing.T) {
		_, err := GetSyncTargetProfileNames("invalid", "")
		if err == nil {
			t.Error("expected error for invalid mode")
		}
	})
}

func TestGetReviewDays(t *testing.T) {
	dir := t.TempDir()
	restore := setEnv("AL_HOME", dir)
	defer restore()

	path := filepath.Join(dir, "profiles.json")
	content := `{"profiles":[
		{"name":"no_review"},
		{"name":"with_review","review_days":30},
		{"name":"zero_review","review_days":0}
	]}`
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	days, has, err := GetReviewDays("no_review")
	if err != nil {
		t.Fatal(err)
	}
	if has {
		t.Error("no_review should not have review")
	}

	days, has, err = GetReviewDays("with_review")
	if err != nil {
		t.Fatal(err)
	}
	if !has || days != 30 {
		t.Errorf("with_review: has=%v days=%d", has, days)
	}

	_, has, _ = GetReviewDays("zero_review")
	if has {
		t.Error("zero_review should not have review")
	}

	_, has, err = GetReviewDays("nonexistent")
	if err != nil || has {
		t.Errorf("nonexistent: err=%v has=%v", err, has)
	}
}

func TestSortProfilesForDisplay(t *testing.T) {
	tests := []struct {
		name  string
		input []ProfileConfig
		want  []string // profile names in expected order
	}{
		{
			name:  "empty",
			input: []ProfileConfig{},
			want:  []string{},
		},
		{
			name: "single profile",
			input: []ProfileConfig{
				{Name: "core"},
			},
			want: []string{"core"},
		},
		{
			name: "core before others",
			input: []ProfileConfig{
				{Name: "work"},
				{Name: "core"},
				{Name: "private"},
			},
			want: []string{"core", "private", "work"},
		},
		{
			name: "stable before trial within same group",
			input: []ProfileConfig{
				{Name: "core.trial"},
				{Name: "core"},
			},
			want: []string{"core", "core.trial"},
		},
		{
			name: "complete example from issue",
			input: []ProfileConfig{
				{Name: "work.trial"},
				{Name: "private"},
				{Name: "core.trial"},
				{Name: "work"},
				{Name: "private.trial"},
				{Name: "core"},
			},
			want: []string{"core", "core.trial", "private", "private.trial", "work", "work.trial"},
		},
		{
			name: "alphabetical order for non-core groups",
			input: []ProfileConfig{
				{Name: "zebra"},
				{Name: "alpha"},
				{Name: "core"},
				{Name: "beta"},
			},
			want: []string{"core", "alpha", "beta", "zebra"},
		},
		{
			name: "multiple stages within group",
			input: []ProfileConfig{
				{Name: "work.trial"},
				{Name: "work.stable"},
				{Name: "work"},
			},
			want: []string{"work", "work.stable", "work.trial"},
		},
		{
			name: "case insensitive sorting",
			input: []ProfileConfig{
				{Name: "Work"},
				{Name: "alpha"},
				{Name: "Beta"},
			},
			want: []string{"alpha", "Beta", "Work"},
		},
		{
			name: "case insensitive stage comparison within same group",
			input: []ProfileConfig{
				{Name: "work.Trial"},
				{Name: "work.Stable"},
				{Name: "work"},
			},
			want: []string{"work", "work.Stable", "work.Trial"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SortProfilesForDisplay(tt.input)
			if len(got) != len(tt.want) {
				t.Fatalf("SortProfilesForDisplay() length = %d, want %d", len(got), len(tt.want))
			}
			for i, want := range tt.want {
				if got[i].Name != want {
					t.Errorf("SortProfilesForDisplay()[%d].Name = %q, want %q", i, got[i].Name, want)
				}
			}
		})
	}
}

func TestSetProfileReviewDays(t *testing.T) {
	dir := t.TempDir()
	restore := setEnv("AL_HOME", dir)
	defer restore()

	if err := EnsureConfigDir(); err != nil {
		t.Fatal(err)
	}

	days30 := 30
	if err := SaveProfilesConfig(&ProfilesConfig{
		Profiles: []ProfileConfig{
			{Name: "trial", Stage: "trial", ReviewDays: &days30},
			{Name: "stable", Stage: "stable"},
		},
	}); err != nil {
		t.Fatal(err)
	}

	// Set review_days on a profile that already has one
	if err := SetProfileReviewDays("trial", 1); err != nil {
		t.Fatalf("SetProfileReviewDays: %v", err)
	}
	days, hasReview, err := GetReviewDays("trial")
	if err != nil {
		t.Fatal(err)
	}
	if !hasReview || days != 1 {
		t.Errorf("Expected review_days=1 after set, got hasReview=%v days=%d", hasReview, days)
	}

	// Set review_days on a profile that had none
	if err := SetProfileReviewDays("stable", 7); err != nil {
		t.Fatalf("SetProfileReviewDays stable: %v", err)
	}
	days, hasReview, err = GetReviewDays("stable")
	if err != nil {
		t.Fatal(err)
	}
	if !hasReview || days != 7 {
		t.Errorf("Expected review_days=7 after set on stable, got hasReview=%v days=%d", hasReview, days)
	}

	// Clear review_days by setting to 0
	if err := SetProfileReviewDays("trial", 0); err != nil {
		t.Fatalf("SetProfileReviewDays clear: %v", err)
	}
	_, hasReview, err = GetReviewDays("trial")
	if err != nil {
		t.Fatal(err)
	}
	if hasReview {
		t.Error("Expected review_days to be cleared after setting to 0")
	}

	// Error on non-existent profile
	if err := SetProfileReviewDays("nonexistent", 5); err == nil {
		t.Error("Expected error for non-existent profile")
	}
}
