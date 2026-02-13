package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/kkato1030/al/internal/config"
	"github.com/kkato1030/al/internal/logger"
	"github.com/kkato1030/al/internal/output"
	"github.com/spf13/cobra"
)

// NewLogsCmd creates the logs command
func NewLogsCmd() *cobra.Command {
	var listOnly bool
	var count int

	cmd := &cobra.Command{
		Use:   "logs [file]",
		Short: "View or list execution logs",
		Long:  "View execution logs from sync and upgrade operations. Without arguments, opens the most recent log. Use --list to show available logs.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			configDir, err := config.GetConfigDir()
			if err != nil {
				return fmt.Errorf("failed to get config directory: %w", err)
			}

			logsDir := logger.GetLogsDir(configDir)

			if listOnly {
				return listLogs(logsDir, count)
			}

			// If specific log file is provided
			if len(args) > 0 {
				return openLog(logsDir, args[0])
			}

			// Open most recent log
			return openRecentLog(logsDir)
		},
	}

	cmd.Flags().BoolVarP(&listOnly, "list", "l", false, "List available logs")
	cmd.Flags().IntVarP(&count, "count", "n", 10, "Number of logs to list (with --list)")

	return cmd
}

func listLogs(logsDir string, count int) error {
	logs, err := logger.ListLogs(logsDir)
	if err != nil {
		return fmt.Errorf("failed to list logs: %w", err)
	}

	if len(logs) == 0 {
		fmt.Println("No logs found")
		return nil
	}

	fmt.Printf("Recent logs (showing up to %d):\n\n", count)

	displayed := 0
	for _, log := range logs {
		if displayed >= count {
			break
		}

		logPath := filepath.Join(logsDir, log)
		info, err := os.Stat(logPath)
		if err != nil {
			continue
		}

		fmt.Printf("  %s  (%s)\n", log, formatSize(info.Size()))
		displayed++
	}

	fmt.Printf("\nLogs location: %s\n", logsDir)
	fmt.Println("Use 'al logs <filename>' to view a specific log")

	return nil
}

func openLog(logsDir, filename string) error {
	logPath := filepath.Join(logsDir, filename)

	// Check if file exists
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		return fmt.Errorf("log file not found: %s", filename)
	}

	return openFile(logPath)
}

func openRecentLog(logsDir string) error {
	logs, err := logger.ListLogs(logsDir)
	if err != nil {
		return fmt.Errorf("failed to list logs: %w", err)
	}

	if len(logs) == 0 {
		return fmt.Errorf("no logs found")
	}

	logPath := filepath.Join(logsDir, logs[0])
	output.Info("Opening most recent log: %s", logs[0])
	return openFile(logPath)
}

func openFile(path string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", path)
	case "linux":
		cmd = exec.Command("xdg-open", path)
	default:
		// Fall back to printing the content
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read log file: %w", err)
		}
		fmt.Println(string(content))
		return nil
	}

	if err := cmd.Run(); err != nil {
		// If opening fails, try to display the content
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			return fmt.Errorf("failed to open log file: %w (and failed to read: %v)", err, readErr)
		}
		fmt.Println(string(content))
	}

	return nil
}

func formatSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	divisor, exponent := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		divisor *= unit
		exponent++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(divisor), "KMGTPE"[exponent])
}
