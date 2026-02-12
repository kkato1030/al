package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
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
