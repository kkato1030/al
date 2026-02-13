package lock

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/kkato1030/al/internal/config"
)

const lockFileName = "lock"

// Lock represents a file-based lock
type Lock struct {
	path      string
	acquired  bool
	signalCh  chan os.Signal
	cleanupCh chan struct{}
}

// New creates a new lock instance
func New() (*Lock, error) {
	configDir, err := config.GetConfigDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get config directory: %w", err)
	}

	return &Lock{
		path:      filepath.Join(configDir, lockFileName),
		acquired:  false,
		signalCh:  make(chan os.Signal, 1),
		cleanupCh: make(chan struct{}),
	}, nil
}

// Acquire attempts to acquire the lock
// If force is true, it will remove any existing lock and acquire a new one
// Returns an error if the lock is already held (and force is false)
func (l *Lock) Acquire(force bool) error {
	// Check if lock already exists
	if _, err := os.Stat(l.path); err == nil {
		if !force {
			// Lock exists and force is not set
			lockInfo, err := l.readLockInfo()
			if err != nil {
				return fmt.Errorf("another al sync/upgrade operation is in progress (lock file: %s)", l.path)
			}
			return fmt.Errorf("another al sync/upgrade operation is in progress (started at %s, lock file: %s)\nUse --force to override this lock", lockInfo, l.path)
		}
		// Force is set, remove existing lock
		if err := os.Remove(l.path); err != nil {
			return fmt.Errorf("failed to remove existing lock: %w", err)
		}
	}

	// Create lock file with timestamp
	lockContent := fmt.Sprintf("pid=%d\ntime=%s\n", os.Getpid(), time.Now().Format(time.RFC3339))
	if err := os.WriteFile(l.path, []byte(lockContent), 0644); err != nil {
		return fmt.Errorf("failed to create lock file: %w", err)
	}

	l.acquired = true

	// Setup signal handler for cleanup on interrupt
	l.setupSignalHandler()

	return nil
}

// Release removes the lock file
func (l *Lock) Release() error {
	if !l.acquired {
		return nil
	}

	// Stop signal handler
	close(l.cleanupCh)

	if err := os.Remove(l.path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove lock file: %w", err)
	}

	l.acquired = false
	return nil
}

// IsLocked checks if a lock file exists
func (l *Lock) IsLocked() bool {
	_, err := os.Stat(l.path)
	return err == nil
}

// setupSignalHandler sets up a handler to clean up the lock on interrupt
func (l *Lock) setupSignalHandler() {
	signal.Notify(l.signalCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		select {
		case <-l.signalCh:
			// Received interrupt signal, cleanup lock
			l.Release()
			os.Exit(130) // Exit code 130 for SIGINT
		case <-l.cleanupCh:
			// Normal cleanup, stop listening
			signal.Stop(l.signalCh)
			return
		}
	}()
}

// readLockInfo reads the timestamp from the lock file
func (l *Lock) readLockInfo() (string, error) {
	content, err := os.ReadFile(l.path)
	if err != nil {
		return "", err
	}

	// Try to parse timestamp from content
	// Format: "pid=<pid>\ntime=<timestamp>\n"
	contentStr := string(content)
	if len(contentStr) == 0 {
		return "unknown time", nil
	}

	// Split by newline and find the time= line
	lines := []string{}
	start := 0
	for i := 0; i < len(contentStr); i++ {
		if contentStr[i] == '\n' {
			lines = append(lines, contentStr[start:i])
			start = i + 1
		}
	}
	if start < len(contentStr) {
		lines = append(lines, contentStr[start:])
	}

	for _, line := range lines {
		if len(line) > 5 && line[:5] == "time=" {
			return line[5:], nil
		}
	}

	// Fallback: return first part of content
	if len(contentStr) > 50 {
		return contentStr[:50], nil
	}
	return contentStr, nil
}
