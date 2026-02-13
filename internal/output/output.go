package output

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/briandowns/spinner"
)

// IsDebugMode returns true if AL_DEBUG environment variable is set
func IsDebugMode() bool {
	return os.Getenv("AL_DEBUG") != ""
}

// DebugLog prints debug messages if AL_DEBUG environment variable is set
func DebugLog(format string, args ...interface{}) {
	if IsDebugMode() {
		timestamp := time.Now().Format("15:04:05.000")
		fmt.Fprintf(os.Stderr, "[DEBUG %s] ", timestamp)
		fmt.Fprintf(os.Stderr, format+"\n", args...)
	}
}

// Info prints an informational message to stdout
func Info(format string, args ...interface{}) {
	fmt.Fprintf(os.Stdout, format+"\n", args...)
}

// Success prints a success message with a checkmark to stdout
func Success(format string, args ...interface{}) {
	fmt.Fprintf(os.Stdout, "✓ "+format+"\n", args...)
}

// Warning prints a warning message
func Warning(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "⚠️  "+format+"\n", args...)
}

// Error prints an error message
func Error(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "✗ "+format+"\n", args...)
}

// ToolOutput prints output from external tools with a prefix
// Only shown when AL_DEBUG is enabled
func ToolOutput(tool string, format string, args ...interface{}) {
	if IsDebugMode() {
		prefix := fmt.Sprintf("[%s] ", tool)
		fmt.Fprintf(os.Stderr, prefix+format+"\n", args...)
	}
}

// Spinner represents a loading spinner for long-running operations
type Spinner struct {
	spinner *spinner.Spinner
}

// NewSpinner creates a new spinner with the given message
func NewSpinner(message string) *Spinner {
	s := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
	s.Suffix = " " + message
	s.Writer = os.Stderr
	return &Spinner{spinner: s}
}

// Start starts the spinner
func (s *Spinner) Start() {
	if !IsDebugMode() {
		s.spinner.Start()
	}
}

// Stop stops the spinner
func (s *Spinner) Stop() {
	if !IsDebugMode() {
		s.spinner.Stop()
	}
}

// Update updates the spinner message
func (s *Spinner) Update(message string) {
	if !IsDebugMode() {
		s.spinner.Suffix = " " + message
	}
}

// DiscardWriter is an io.Writer that discards all output
type DiscardWriter struct{}

func (d DiscardWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

// PrefixWriter wraps an io.Writer and adds a prefix to each line
type PrefixWriter struct {
	writer io.Writer
	prefix string
}

// NewPrefixWriter creates a new PrefixWriter
func NewPrefixWriter(w io.Writer, prefix string) *PrefixWriter {
	return &PrefixWriter{
		writer: w,
		prefix: prefix,
	}
}

func (p *PrefixWriter) Write(data []byte) (n int, err error) {
	_, err = p.writer.Write(append([]byte(p.prefix), data...))
	if err != nil {
		return 0, err
	}
	// Return the number of bytes consumed from input, not including prefix
	return len(data), nil
}

// GetToolOutputWriter returns the appropriate writer for tool output
// In debug mode, returns stderr with prefix. Otherwise, discards output.
func GetToolOutputWriter(tool string) io.Writer {
	if IsDebugMode() {
		return NewPrefixWriter(os.Stderr, fmt.Sprintf("[%s] ", tool))
	}
	return DiscardWriter{}
}
