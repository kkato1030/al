package brewfile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseFile(t *testing.T) {
	dir := t.TempDir()

	tests := []struct {
		name     string
		content  string
		wantErr  bool
		wantN    int
		wantSkip int
		check    func(t *testing.T, r *ParseResult)
	}{
		{
			name:     "empty file",
			content:  "",
			wantN:    0,
			wantSkip: 0,
		},
		{
			name: "comments and empty lines only",
			content: `
# comment
   
# another
`,
			wantN:    0,
			wantSkip: 0,
		},
		{
			name:    "tap single",
			content: `tap "homebrew/cask"`,
			wantN:   1,
			check: func(t *testing.T, r *ParseResult) {
				if len(r.Entries) != 1 {
					t.Fatalf("entries: got %d", len(r.Entries))
				}
				e := r.Entries[0]
				if e.Provider != "brew" || e.ID != "tap:homebrew/cask" || e.Name != "homebrew/cask" {
					t.Errorf("entry: provider=%q id=%q name=%q", e.Provider, e.ID, e.Name)
				}
			},
		},
		{
			name:    "brew formula",
			content: `brew "ruby"`,
			wantN:   1,
			check: func(t *testing.T, r *ParseResult) {
				e := r.Entries[0]
				if e.Provider != "brew" || e.ID != "formula:ruby" || e.Name != "ruby" {
					t.Errorf("entry: provider=%q id=%q name=%q", e.Provider, e.ID, e.Name)
				}
			},
		},
		{
			name:    "brew formula with version",
			content: `brew "postgresql@16"`,
			wantN:   1,
			check: func(t *testing.T, r *ParseResult) {
				e := r.Entries[0]
				if e.ID != "formula:postgresql@16" {
					t.Errorf("id: got %q", e.ID)
				}
			},
		},
		{
			name:    "cask",
			content: `cask "firefox"`,
			wantN:   1,
			check: func(t *testing.T, r *ParseResult) {
				e := r.Entries[0]
				if e.Provider != "brew" || e.ID != "cask:firefox" || e.Name != "firefox" {
					t.Errorf("entry: provider=%q id=%q name=%q", e.Provider, e.ID, e.Name)
				}
			},
		},
		{
			name:    "mas with id",
			content: `mas "Slack", id: 803453959`,
			wantN:   1,
			check: func(t *testing.T, r *ParseResult) {
				e := r.Entries[0]
				if e.Provider != "mas" || e.ID != "803453959" || e.Name != "Slack" {
					t.Errorf("entry: provider=%q id=%q name=%q", e.Provider, e.ID, e.Name)
				}
			},
		},
		{
			name:     "vscode skipped",
			content:  `vscode "editorconfig.editorconfig"`,
			wantN:    0,
			wantSkip: 1,
			check: func(t *testing.T, r *ParseResult) {
				if len(r.Skipped) != 1 || r.Skipped[0].Reason != "vscode" {
					t.Errorf("skipped: %+v", r.Skipped)
				}
			},
		},
		{
			name:     "mas without id skipped",
			content:  `mas "Some App"`,
			wantN:    0,
			wantSkip: 1,
			check: func(t *testing.T, r *ParseResult) {
				if len(r.Skipped) != 1 {
					t.Errorf("skipped: got %d", len(r.Skipped))
				}
			},
		},
		{
			name: "mixed",
			content: `tap "homebrew/cask"
brew "ruby"
# comment
cask "firefox"
mas "Slack", id: 803453959
vscode "ext"
`,
			wantN:    4,
			wantSkip: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(dir, "Brewfile")
			if err := os.WriteFile(path, []byte(tt.content), 0644); err != nil {
				t.Fatal(err)
			}
			r, err := ParseFile(path)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseFile() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}
			if len(r.Entries) != tt.wantN {
				t.Errorf("entries: got %d, want %d", len(r.Entries), tt.wantN)
			}
			if len(r.Skipped) != tt.wantSkip {
				t.Errorf("skipped: got %d, want %d", len(r.Skipped), tt.wantSkip)
			}
			if tt.check != nil {
				tt.check(t, r)
			}
		})
	}
}

func TestParseFile_NotFound(t *testing.T) {
	_, err := ParseFile("/nonexistent/path/Brewfile")
	if err == nil {
		t.Error("ParseFile: expected error for nonexistent file")
	}
}

func TestParseFile_Testdata(t *testing.T) {
	// Use testdata/Brewfile if present (relative to package dir or repo root)
	paths := []string{
		"testdata/Brewfile",
		"../../testdata/Brewfile",
	}
	var path string
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			path = p
			break
		}
	}
	if path == "" {
		t.Skip("testdata/Brewfile not found")
	}
	r, err := ParseFile(path)
	if err != nil {
		t.Fatal(err)
	}
	// testdata/Brewfile has: tap, brew x2, cask, mas, and vscode (skipped)
	if len(r.Entries) < 4 {
		t.Errorf("entries: got %d, want at least 4", len(r.Entries))
	}
	if len(r.Skipped) < 1 {
		t.Errorf("skipped: got %d, want at least 1", len(r.Skipped))
	}
}

func TestResolveBrewfilePath(t *testing.T) {
	dir := t.TempDir()
	customPath := filepath.Join(dir, "MyBrewfile")
	if err := os.WriteFile(customPath, []byte("brew \"foo\""), 0644); err != nil {
		t.Fatal(err)
	}

	t.Run("user path non-empty", func(t *testing.T) {
		got, err := ResolveBrewfilePath(customPath)
		if err != nil {
			t.Fatal(err)
		}
		if got != customPath {
			t.Errorf("got %q, want %q", got, customPath)
		}
	})

	t.Run("user path empty prefers current Brewfile", func(t *testing.T) {
		// Run from dir that has Brewfile
		withBrewfile := filepath.Join(dir, "sub")
		if err := os.MkdirAll(withBrewfile, 0755); err != nil {
			t.Fatal(err)
		}
		brewfilePath := filepath.Join(withBrewfile, "Brewfile")
		if err := os.WriteFile(brewfilePath, []byte("brew \"a\""), 0644); err != nil {
			t.Fatal(err)
		}
		origWd, _ := os.Getwd()
		defer os.Chdir(origWd)
		if err := os.Chdir(withBrewfile); err != nil {
			t.Fatal(err)
		}
		got, err := ResolveBrewfilePath("")
		if err != nil {
			t.Fatal(err)
		}
		if got != "Brewfile" {
			t.Errorf("got %q, want Brewfile", got)
		}
	})
}
