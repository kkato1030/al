package cmd

import (
	"fmt"
	"os"
	"strings"

	packagecmd "github.com/kkato1030/al/cmd/package"
	providercmd "github.com/kkato1030/al/cmd/provider"
	"github.com/kkato1030/al/internal/config"
	"github.com/kkato1030/al/internal/lock"
	"github.com/kkato1030/al/internal/logger"
	"github.com/spf13/cobra"
)

// NewUpgradeCmd creates the upgrade command
func NewUpgradeCmd() *cobra.Command {
	var yes bool
	var force bool

	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade all providers and packages",
		Long:  "Upgrade all providers and packages. This is equivalent to running 'al provider upgrade' followed by 'al package upgrade'. Use --force to override any existing lock.",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Create logger for upgrade operation
			var log *logger.Logger
			configDir, err := config.GetConfigDir()
			if err == nil {
				logsDir := logger.GetLogsDir(configDir)
				log, err = logger.New(logsDir, "al upgrade")
				if err != nil {
					// Log creation failed, but don't fail the upgrade
					fmt.Fprintf(os.Stderr, "Warning: failed to create log file: %v\n", err)
				}
			}

			err = runUpgrade(yes, force, log)

			if log != nil {
				log.Close()
			}

			return err
		},
	}

	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation prompt")
	cmd.Flags().BoolVar(&force, "force", false, "Override any existing lock from a previous sync/upgrade operation")

	return cmd
}

func runUpgrade(yes bool, force bool, log *logger.Logger) error {
	// Acquire lock
	lockInstance, err := lock.New()
	if err != nil {
		return fmt.Errorf("failed to create lock: %w", err)
	}

	if err := lockInstance.Acquire(force); err != nil {
		return err
	}
	defer lockInstance.Release()

	if log != nil {
		log.WriteString("Lock acquired\n")
	}

	// Helper to log and print
	logPrintf := func(format string, args ...interface{}) {
		msg := fmt.Sprintf(format, args...)
		fmt.Print(msg)
		if log != nil {
			log.WriteString(msg)
		}
	}

	// Ask for confirmation
	if !yes {
		logPrintf("This will upgrade all providers and packages.\n")
		logPrintf("This is equivalent to:\n")
		logPrintf("  1. al provider upgrade\n")
		logPrintf("  2. al package upgrade\n")
		logPrintf("\nDo you want to continue? [y/N]: ")
		var response string
		fmt.Scanln(&response)
		if log != nil {
			log.WriteString(response + "\n")
		}
		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			logPrintf("Upgrade cancelled.\n")
			return nil
		}
	}

	// Upgrade all providers
	logPrintf("\n")
	if err := providercmd.RunProviderUpgradeAll(true); err != nil {
		logPrintf("\nError upgrading providers: %v\n", err)
		// Continue to package upgrade even if provider upgrade fails
	}

	// Upgrade all packages
	logPrintf("\n")
	if err := packagecmd.RunPackageUpgradeAll(true); err != nil {
		if log != nil {
			log.WriteString(fmt.Sprintf("error upgrading packages: %v\n", err))
		}
		return fmt.Errorf("error upgrading packages: %w", err)
	}

	logPrintf("\n✓ All upgrades completed\n")
	return nil
}
