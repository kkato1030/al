package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/kkato1030/al/internal/config"
)

func TestDiffJSONOutput(t *testing.T) {
	// Set JSON mode
	jsonOutput = true
	defer func() { jsonOutput = false }()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Create a simple diff result
	result := diffResult{
		additions: []packageDiff{
			{provider: "brew", name: "jq", id: "formula:jq"},
		},
		removals: []packageDiff{
			{provider: "brew", name: "wget", id: "formula:wget"},
		},
		upgrades: []packageDiff{
			{provider: "brew", name: "git", id: "formula:git"},
		},
	}

	// Display the diff
	displayDiff(result)

	// Restore stdout and read output
	w.Close()
	os.Stdout = oldStdout
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Parse JSON output
	var jsonOut DiffJSONOutput
	if err := json.Unmarshal([]byte(output), &jsonOut); err != nil {
		t.Fatalf("Failed to parse JSON output: %v\nOutput: %s", err, output)
	}

	// Verify the output
	if !jsonOut.HasDrift {
		t.Error("Expected HasDrift to be true")
	}

	if len(jsonOut.Additions) != 1 {
		t.Errorf("Expected 1 addition, got %d", len(jsonOut.Additions))
	} else if jsonOut.Additions[0].Name != "jq" {
		t.Errorf("Expected addition 'jq', got '%s'", jsonOut.Additions[0].Name)
	}

	if len(jsonOut.Removals) != 1 {
		t.Errorf("Expected 1 removal, got %d", len(jsonOut.Removals))
	} else if jsonOut.Removals[0].Name != "wget" {
		t.Errorf("Expected removal 'wget', got '%s'", jsonOut.Removals[0].Name)
	}

	if len(jsonOut.Upgrades) != 1 {
		t.Errorf("Expected 1 upgrade, got %d", len(jsonOut.Upgrades))
	} else if jsonOut.Upgrades[0].Name != "git" {
		t.Errorf("Expected upgrade 'git', got '%s'", jsonOut.Upgrades[0].Name)
	}
}

func TestDiffJSONOutputNoDrift(t *testing.T) {
	// Set JSON mode
	jsonOutput = true
	defer func() { jsonOutput = false }()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Create empty diff result
	result := diffResult{
		additions: []packageDiff{},
		removals:  []packageDiff{},
		upgrades:  []packageDiff{},
	}

	// Display the diff
	displayDiff(result)

	// Restore stdout and read output
	w.Close()
	os.Stdout = oldStdout
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Parse JSON output
	var jsonOut DiffJSONOutput
	if err := json.Unmarshal([]byte(output), &jsonOut); err != nil {
		t.Fatalf("Failed to parse JSON output: %v\nOutput: %s", err, output)
	}

	// Verify the output
	if jsonOut.HasDrift {
		t.Error("Expected HasDrift to be false")
	}

	if len(jsonOut.Additions) != 0 {
		t.Errorf("Expected 0 additions, got %d", len(jsonOut.Additions))
	}
}

func TestDoctorJSONOutput(t *testing.T) {
	// Set JSON mode
	jsonOutput = true
	defer func() { jsonOutput = false }()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Create test results
	results := []CheckResult{
		{Status: StatusOK, Message: "test ok"},
		{Status: StatusWarn, Message: "test warn"},
		{Status: StatusError, Message: "test error"},
	}

	// Print results
	printResults(results)

	// Restore stdout and read output
	w.Close()
	os.Stdout = oldStdout
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Parse JSON output
	var jsonOut DoctorJSONOutput
	if err := json.Unmarshal([]byte(output), &jsonOut); err != nil {
		t.Fatalf("Failed to parse JSON output: %v\nOutput: %s", err, output)
	}

	// Verify the output
	if !jsonOut.HasError {
		t.Error("Expected HasError to be true")
	}

	if jsonOut.Summary.OK != 1 {
		t.Errorf("Expected 1 OK, got %d", jsonOut.Summary.OK)
	}

	if jsonOut.Summary.Warnings != 1 {
		t.Errorf("Expected 1 warning, got %d", jsonOut.Summary.Warnings)
	}

	if jsonOut.Summary.Errors != 1 {
		t.Errorf("Expected 1 error, got %d", jsonOut.Summary.Errors)
	}

	if len(jsonOut.Checks) != 3 {
		t.Errorf("Expected 3 checks, got %d", len(jsonOut.Checks))
	}
}

func TestSyncPlanJSONOutput(t *testing.T) {
	// Set JSON mode
	jsonOutput = true
	defer func() { jsonOutput = false }()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Create test data
	syncTargetNames := []string{"core.stable"}
	toInstall := []config.PackageConfig{
		{ID: "formula:jq", Name: "jq", Provider: "brew", Profile: "core.stable"},
	}
	toUpgrade := []config.PackageConfig{
		{ID: "formula:git", Name: "git", Provider: "brew", Profile: "core.stable"},
	}
	manualPackages := []config.PackageConfig{
		{ID: "manual-1", Name: "manual-tool", Provider: "manual", Profile: "core.stable"},
	}
	providersToInstall := []string{"mas"}
	links := []config.LinkEntry{
		{Name: "testlink", Manifest: &config.LinkManifest{UserPath: "/home/user/.config"}},
	}
	hasBootstrap := true

	// Output sync plan
	outputSyncPlan(syncTargetNames, toInstall, toUpgrade, manualPackages, providersToInstall, links, hasBootstrap)

	// Restore stdout and read output
	w.Close()
	os.Stdout = oldStdout
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Parse JSON output
	var jsonOut SyncPlanJSONOutput
	if err := json.Unmarshal([]byte(output), &jsonOut); err != nil {
		t.Fatalf("Failed to parse JSON output: %v\nOutput: %s", err, output)
	}

	// Verify the output
	if jsonOut.Summary.Providers != 1 {
		t.Errorf("Expected 1 provider, got %d", jsonOut.Summary.Providers)
	}

	if jsonOut.Summary.Installations != 1 {
		t.Errorf("Expected 1 installation, got %d", jsonOut.Summary.Installations)
	}

	if jsonOut.Summary.Upgrades != 1 {
		t.Errorf("Expected 1 upgrade, got %d", jsonOut.Summary.Upgrades)
	}

	if jsonOut.Summary.ManualPackages != 1 {
		t.Errorf("Expected 1 manual package, got %d", jsonOut.Summary.ManualPackages)
	}

	if jsonOut.Summary.Links != 1 {
		t.Errorf("Expected 1 link, got %d", jsonOut.Summary.Links)
	}

	if !jsonOut.Summary.Bootstrap {
		t.Error("Expected Bootstrap to be true")
	}

	// Verify actions
	expectedActionCount := 1 + 1 + 1 + 1 + 1 + 1 // provider + install + upgrade + manual + link + bootstrap
	if len(jsonOut.Actions) != expectedActionCount {
		t.Errorf("Expected %d actions, got %d", expectedActionCount, len(jsonOut.Actions))
	}

	// Check action types
	foundInstallProvider := false
	foundInstall := false
	foundUpgrade := false
	foundManual := false
	foundLink := false
	foundBootstrap := false

	for _, action := range jsonOut.Actions {
		switch action.Type {
		case "install_provider":
			foundInstallProvider = true
		case "install":
			foundInstall = true
		case "upgrade":
			foundUpgrade = true
		case "manual":
			foundManual = true
		case "link":
			foundLink = true
		case "bootstrap":
			foundBootstrap = true
		}
	}

	if !foundInstallProvider {
		t.Error("Expected to find install_provider action")
	}
	if !foundInstall {
		t.Error("Expected to find install action")
	}
	if !foundUpgrade {
		t.Error("Expected to find upgrade action")
	}
	if !foundManual {
		t.Error("Expected to find manual action")
	}
	if !foundLink {
		t.Error("Expected to find link action")
	}
	if !foundBootstrap {
		t.Error("Expected to find bootstrap action")
	}
}
