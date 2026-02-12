package logger

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	tmpDir := t.TempDir()
	logsDir := filepath.Join(tmpDir, "logs")

	l, err := New(logsDir, "al sync")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer l.Close()

	// Check that logs directory was created
	if _, err := os.Stat(logsDir); os.IsNotExist(err) {
		t.Errorf("logs directory was not created")
	}

	// Check that log file was created
	if l.file == nil {
		t.Errorf("log file was not created")
	}
}

func TestLogger_Write(t *testing.T) {
	tmpDir := t.TempDir()
	logsDir := filepath.Join(tmpDir, "logs")

	l, err := New(logsDir, "al sync")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// Write some data
	testData := "test log message\n"
	n, err := l.Write([]byte(testData))
	if err != nil {
		t.Errorf("Write() error = %v", err)
	}
	if n == 0 {
		t.Errorf("Write() wrote 0 bytes")
	}

	l.Close()

	// Read the log file and verify content
	files, err := os.ReadDir(logsDir)
	if err != nil || len(files) == 0 {
		t.Fatalf("failed to read log files: %v", err)
	}

	logPath := filepath.Join(logsDir, files[0].Name())
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "Command: al sync") {
		t.Errorf("log file does not contain command")
	}
	if !strings.Contains(contentStr, testData) {
		t.Errorf("log file does not contain written data")
	}
	if !strings.Contains(contentStr, "Started:") {
		t.Errorf("log file does not contain start timestamp")
	}
	if !strings.Contains(contentStr, "Finished:") {
		t.Errorf("log file does not contain finish timestamp")
	}
}

func TestGetLogsDir(t *testing.T) {
	configDir := "/home/user/.al"
	want := "/home/user/.al/logs"
	got := GetLogsDir(configDir)
	if got != want {
		t.Errorf("GetLogsDir() = %q, want %q", got, want)
	}
}

func TestListLogs(t *testing.T) {
	tmpDir := t.TempDir()
	logsDir := filepath.Join(tmpDir, "logs")
	os.MkdirAll(logsDir, 0755)

	// Create some test log files
	testFiles := []string{
		"20240101-120000.log",
		"20240102-130000.log",
		"20240103-140000.log",
		"not-a-log.txt",
	}

	for _, name := range testFiles {
		path := filepath.Join(logsDir, name)
		if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
	}

	// Give files slightly different mod times
	time.Sleep(10 * time.Millisecond)

	logs, err := ListLogs(logsDir)
	if err != nil {
		t.Fatalf("ListLogs() error = %v", err)
	}

	// Should only return .log files
	if len(logs) != 3 {
		t.Errorf("ListLogs() returned %d files, want 3", len(logs))
	}

	// Should be sorted newest first (reverse lexical for YYYYMMDD-HHMMSS format)
	if logs[0] != "20240103-140000.log" {
		t.Errorf("ListLogs()[0] = %q, want 20240103-140000.log", logs[0])
	}
	if logs[2] != "20240101-120000.log" {
		t.Errorf("ListLogs()[2] = %q, want 20240101-120000.log", logs[2])
	}
}

func TestListLogs_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	logsDir := filepath.Join(tmpDir, "logs")

	logs, err := ListLogs(logsDir)
	if err != nil {
		t.Fatalf("ListLogs() error = %v", err)
	}
	if len(logs) != 0 {
		t.Errorf("ListLogs() on non-existent directory should return empty slice, got %v", logs)
	}
}
