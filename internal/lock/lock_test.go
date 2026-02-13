package lock

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	// Set up temporary AL_HOME
	tmpDir := t.TempDir()
	oldHome := os.Getenv("AL_HOME")
	os.Setenv("AL_HOME", tmpDir)
	defer func() {
		if oldHome == "" {
			os.Unsetenv("AL_HOME")
		} else {
			os.Setenv("AL_HOME", oldHome)
		}
	}()

	lock, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	expectedPath := filepath.Join(tmpDir, lockFileName)
	if lock.path != expectedPath {
		t.Errorf("lock.path = %q, want %q", lock.path, expectedPath)
	}

	if lock.acquired {
		t.Error("lock.acquired should be false initially")
	}
}

func TestAcquire_Success(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("AL_HOME")
	os.Setenv("AL_HOME", tmpDir)
	defer func() {
		if oldHome == "" {
			os.Unsetenv("AL_HOME")
		} else {
			os.Setenv("AL_HOME", oldHome)
		}
	}()

	lock, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	err = lock.Acquire(false)
	if err != nil {
		t.Fatalf("Acquire() error = %v", err)
	}
	defer lock.Release()

	if !lock.acquired {
		t.Error("lock.acquired should be true after Acquire()")
	}

	// Check that lock file exists
	if _, err := os.Stat(lock.path); os.IsNotExist(err) {
		t.Error("lock file should exist after Acquire()")
	}

	// Check lock file content
	content, err := os.ReadFile(lock.path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if len(content) == 0 {
		t.Error("lock file should not be empty")
	}
}

func TestAcquire_AlreadyLocked(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("AL_HOME")
	os.Setenv("AL_HOME", tmpDir)
	defer func() {
		if oldHome == "" {
			os.Unsetenv("AL_HOME")
		} else {
			os.Setenv("AL_HOME", oldHome)
		}
	}()

	// First lock
	lock1, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if err := lock1.Acquire(false); err != nil {
		t.Fatalf("Acquire() error = %v", err)
	}
	defer lock1.Release()

	// Try to acquire second lock
	lock2, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	err = lock2.Acquire(false)
	if err == nil {
		lock2.Release()
		t.Fatal("Acquire() should return error when lock already exists")
	}
}

func TestAcquire_ForceOverride(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("AL_HOME")
	os.Setenv("AL_HOME", tmpDir)
	defer func() {
		if oldHome == "" {
			os.Unsetenv("AL_HOME")
		} else {
			os.Setenv("AL_HOME", oldHome)
		}
	}()

	// First lock
	lock1, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if err := lock1.Acquire(false); err != nil {
		t.Fatalf("Acquire() error = %v", err)
	}
	// Don't defer Release for lock1, let lock2 override it

	// Try to acquire second lock with force
	lock2, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	err = lock2.Acquire(true)
	if err != nil {
		t.Fatalf("Acquire(force=true) should succeed, got error = %v", err)
	}
	defer lock2.Release()

	if !lock2.acquired {
		t.Error("lock2.acquired should be true after Acquire(force=true)")
	}
}

func TestRelease(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("AL_HOME")
	os.Setenv("AL_HOME", tmpDir)
	defer func() {
		if oldHome == "" {
			os.Unsetenv("AL_HOME")
		} else {
			os.Setenv("AL_HOME", oldHome)
		}
	}()

	lock, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if err := lock.Acquire(false); err != nil {
		t.Fatalf("Acquire() error = %v", err)
	}

	if err := lock.Release(); err != nil {
		t.Fatalf("Release() error = %v", err)
	}

	if lock.acquired {
		t.Error("lock.acquired should be false after Release()")
	}

	// Check that lock file is removed
	if _, err := os.Stat(lock.path); !os.IsNotExist(err) {
		t.Error("lock file should be removed after Release()")
	}
}

func TestIsLocked(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("AL_HOME")
	os.Setenv("AL_HOME", tmpDir)
	defer func() {
		if oldHome == "" {
			os.Unsetenv("AL_HOME")
		} else {
			os.Setenv("AL_HOME", oldHome)
		}
	}()

	lock, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// Initially not locked
	if lock.IsLocked() {
		t.Error("IsLocked() should return false initially")
	}

	// After acquiring
	if err := lock.Acquire(false); err != nil {
		t.Fatalf("Acquire() error = %v", err)
	}
	defer lock.Release()

	if !lock.IsLocked() {
		t.Error("IsLocked() should return true after Acquire()")
	}
}

func TestLockContent(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("AL_HOME")
	os.Setenv("AL_HOME", tmpDir)
	defer func() {
		if oldHome == "" {
			os.Unsetenv("AL_HOME")
		} else {
			os.Setenv("AL_HOME", oldHome)
		}
	}()

	lock, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	before := time.Now()
	if err := lock.Acquire(false); err != nil {
		t.Fatalf("Acquire() error = %v", err)
	}
	defer lock.Release()
	after := time.Now()

	// Read lock file content
	content, err := os.ReadFile(lock.path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	contentStr := string(content)

	// Check that content contains pid
	if !strings.Contains(contentStr, "pid=") {
		t.Error("lock file should contain pid")
	}

	// Check that content contains time
	if !strings.Contains(contentStr, "time=") {
		t.Error("lock file should contain time")
	}

	// Parse time from lock info
	lockInfo, err := lock.readLockInfo()
	if err != nil {
		t.Fatalf("readLockInfo() error = %v", err)
	}

	// Parse the time
	lockTime, err := time.Parse(time.RFC3339, lockInfo)
	if err != nil {
		t.Fatalf("failed to parse lock time %q: %v", lockInfo, err)
	}

	// Check that lock time is between before and after (with some tolerance)
	// Allow 2 seconds of difference to account for time resolution
	if lockTime.Before(before.Add(-2*time.Second)) || lockTime.After(after.Add(2*time.Second)) {
		t.Errorf("lock time %v should be between %v and %v", lockTime, before, after)
	}
}
