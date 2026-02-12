package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

const (
	// DefaultMaxLogs is the default number of log files to keep
	DefaultMaxLogs = 30
)

// Logger captures execution logs
type Logger struct {
	file      *os.File
	mu        sync.Mutex
	startTime time.Time
	command   string
}

// New creates a new logger that writes to a timestamped log file
func New(logsDir, command string) (*Logger, error) {
	// Ensure logs directory exists
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create logs directory: %w", err)
	}

	// Create log file with timestamp
	timestamp := time.Now().Format("20060102-150405")
	logPath := filepath.Join(logsDir, timestamp+".log")

	file, err := os.Create(logPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create log file: %w", err)
	}

	l := &Logger{
		file:      file,
		startTime: time.Now(),
		command:   command,
	}

	// Write header
	l.writeHeader()

	// Rotate old logs (keep only the most recent ones)
	// Do this after creating the new log file, in a goroutine to not block
	go func() {
		if err := RotateLogs(logsDir, DefaultMaxLogs); err != nil {
			// Log rotation errors are non-fatal, just print to stderr
			fmt.Fprintf(os.Stderr, "Warning: log rotation failed: %v\n", err)
		}
	}()

	return l, nil
}

// writeHeader writes the log header
func (l *Logger) writeHeader() {
	l.mu.Lock()
	defer l.mu.Unlock()

	fmt.Fprintf(l.file, "=== al execution log ===\n")
	fmt.Fprintf(l.file, "Command: %s\n", l.command)
	fmt.Fprintf(l.file, "Started: %s\n", l.startTime.Format(time.RFC3339))
	fmt.Fprintf(l.file, "========================\n\n")
}

// Write writes data to the log file
func (l *Logger) Write(p []byte) (n int, err error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Write timestamp and data
	timestamp := time.Now().Format("15:04:05")
	fmt.Fprintf(l.file, "[%s] ", timestamp)
	return l.file.Write(p)
}

// WriteString writes a string to the log file
func (l *Logger) WriteString(s string) (n int, err error) {
	return l.Write([]byte(s))
}

// Close closes the log file
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file != nil {
		// Write footer
		duration := time.Since(l.startTime)
		fmt.Fprintf(l.file, "\n========================\n")
		fmt.Fprintf(l.file, "Finished: %s\n", time.Now().Format(time.RFC3339))
		fmt.Fprintf(l.file, "Duration: %s\n", duration)
		fmt.Fprintf(l.file, "========================\n")

		return l.file.Close()
	}
	return nil
}

// MultiWriter creates a writer that writes to both the logger and another writer (like os.Stdout)
func (l *Logger) MultiWriter(w io.Writer) io.Writer {
	return io.MultiWriter(l, w)
}

// GetLogsDir returns the logs directory path
func GetLogsDir(configDir string) string {
	return filepath.Join(configDir, "logs")
}

// ListLogs returns a list of log files sorted by modification time (newest first)
func ListLogs(logsDir string) ([]string, error) {
	entries, err := os.ReadDir(logsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	var logs []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".log" {
			logs = append(logs, entry.Name())
		}
	}

	// Sort in reverse order (newest first)
	// Since filenames are YYYYMMDD-HHMMSS.log, lexical sorting works
	for i := len(logs)/2 - 1; i >= 0; i-- {
		opp := len(logs) - 1 - i
		logs[i], logs[opp] = logs[opp], logs[i]
	}

	return logs, nil
}

// RotateLogs removes old log files, keeping only the most recent maxLogs files
func RotateLogs(logsDir string, maxLogs int) error {
	if maxLogs <= 0 {
		return fmt.Errorf("maxLogs must be positive, got %d", maxLogs)
	}

	entries, err := os.ReadDir(logsDir)
	if err != nil {
		if os.IsNotExist(err) {
			// Directory doesn't exist, nothing to rotate
			return nil
		}
		return fmt.Errorf("failed to read logs directory: %w", err)
	}

	// Collect log files with their modification times
	type logFile struct {
		name    string
		modTime time.Time
	}
	var logFiles []logFile

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".log" {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue // Skip files we can't stat
		}

		logFiles = append(logFiles, logFile{
			name:    entry.Name(),
			modTime: info.ModTime(),
		})
	}

	// If we have maxLogs or fewer, no rotation needed
	if len(logFiles) <= maxLogs {
		return nil
	}

	// Sort by modification time (newest first)
	sort.Slice(logFiles, func(i, j int) bool {
		return logFiles[i].modTime.After(logFiles[j].modTime)
	})

	// Delete old logs (keep only maxLogs newest)
	for i := maxLogs; i < len(logFiles); i++ {
		logPath := filepath.Join(logsDir, logFiles[i].name)
		if err := os.Remove(logPath); err != nil {
			// Don't fail the entire rotation if one file fails to delete
			fmt.Fprintf(os.Stderr, "Warning: failed to remove old log %s: %v\n", logFiles[i].name, err)
		}
	}

	return nil
}
