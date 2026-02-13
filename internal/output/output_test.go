package output

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestIsDebugMode(t *testing.T) {
	// Save original value
	originalValue := os.Getenv("AL_DEBUG")
	defer os.Setenv("AL_DEBUG", originalValue)

	// Test when AL_DEBUG is not set
	os.Unsetenv("AL_DEBUG")
	if IsDebugMode() {
		t.Error("Expected IsDebugMode() to return false when AL_DEBUG is not set")
	}

	// Test when AL_DEBUG is set
	os.Setenv("AL_DEBUG", "1")
	if !IsDebugMode() {
		t.Error("Expected IsDebugMode() to return true when AL_DEBUG is set")
	}

	// Test when AL_DEBUG is empty string
	os.Setenv("AL_DEBUG", "")
	if IsDebugMode() {
		t.Error("Expected IsDebugMode() to return false when AL_DEBUG is empty")
	}
}

func TestDiscardWriter(t *testing.T) {
	dw := DiscardWriter{}
	testData := []byte("test data")

	n, err := dw.Write(testData)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if n != len(testData) {
		t.Errorf("Expected %d bytes written, got %d", len(testData), n)
	}
}

func TestPrefixWriter(t *testing.T) {
	buf := &bytes.Buffer{}
	pw := NewPrefixWriter(buf, "[test] ")

	testData := []byte("hello world")
	n, err := pw.Write(testData)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if n != len(testData)+len("[test] ") {
		t.Errorf("Expected %d bytes written, got %d", len(testData)+len("[test] "), n)
	}

	output := buf.String()
	if !strings.HasPrefix(output, "[test] ") {
		t.Errorf("Expected output to start with '[test] ', got %q", output)
	}
	if !strings.Contains(output, "hello world") {
		t.Errorf("Expected output to contain 'hello world', got %q", output)
	}
}

func TestGetToolOutputWriter(t *testing.T) {
	// Save original value
	originalValue := os.Getenv("AL_DEBUG")
	defer os.Setenv("AL_DEBUG", originalValue)

	// Test when AL_DEBUG is not set - should return DiscardWriter
	os.Unsetenv("AL_DEBUG")
	writer := GetToolOutputWriter("brew")
	if _, ok := writer.(DiscardWriter); !ok {
		t.Error("Expected DiscardWriter when AL_DEBUG is not set")
	}

	// Test when AL_DEBUG is set - should return PrefixWriter
	os.Setenv("AL_DEBUG", "1")
	writer = GetToolOutputWriter("brew")
	if _, ok := writer.(*PrefixWriter); !ok {
		t.Error("Expected PrefixWriter when AL_DEBUG is set")
	}
}

func TestSpinner(t *testing.T) {
	// Save original value
	originalValue := os.Getenv("AL_DEBUG")
	defer os.Setenv("AL_DEBUG", originalValue)

	// Just test that spinner can be created and started/stopped without panic
	// In debug mode, spinner should not actually spin
	os.Setenv("AL_DEBUG", "1")
	spinner := NewSpinner("test message")
	if spinner == nil {
		t.Fatal("Expected spinner to be created")
	}

	// These should not panic
	spinner.Start()
	spinner.Update("updated message")
	spinner.Stop()

	// Test in non-debug mode
	os.Unsetenv("AL_DEBUG")
	spinner = NewSpinner("test message")
	if spinner == nil {
		t.Fatal("Expected spinner to be created")
	}

	// These should not panic
	spinner.Start()
	spinner.Update("updated message")
	spinner.Stop()
}
