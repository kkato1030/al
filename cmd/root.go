package cmd

import (
	bootstrapcmd "github.com/kkato1030/al/cmd/bootstrap"
	configcmd "github.com/kkato1030/al/cmd/config"
	linkcmd "github.com/kkato1030/al/cmd/link"
	packagecmd "github.com/kkato1030/al/cmd/package"
	"github.com/kkato1030/al/cmd/profile"
	"github.com/kkato1030/al/cmd/provider"
	shellcmd "github.com/kkato1030/al/cmd/shell"
	"github.com/spf13/cobra"
)

var version = "0.1.0"

// jsonOutput is a global flag to enable JSON output
var jsonOutput bool

// SetVersion sets the version string
func SetVersion(v string) {
	version = v
}

// IsJSONOutput returns whether JSON output is enabled
func IsJSONOutput() bool {
	return jsonOutput
}

// NewRootCmd creates the root command
func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:           "al",
		Short:         "Mac Management Tools",
		Long:          "al - Mac Management Tools",
		SilenceErrors: true, // Prevent Cobra from printing errors automatically (we handle it in main.go)
	}

	helpTemplate := `{{with (or .Long .Short)}}{{. | trimTrailingWhitespaces}}

{{end}}{{if or .Runnable .HasSubCommands}}
Usage:
{{if .Runnable}}  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}  {{.UseLine}} [command]{{end}}{{end}}{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`
	rootCmd.SetHelpTemplate(helpTemplate)

	rootCmd.AddCommand(NewVersionCmd())
	rootCmd.AddCommand(NewInitCmd())
	rootCmd.AddCommand(NewBackupCmd())
	rootCmd.AddCommand(NewPullCmd())
	rootCmd.AddCommand(NewSyncCmd())
	rootCmd.AddCommand(NewDiffCmd())
	rootCmd.AddCommand(NewUpdateCmd())
	rootCmd.AddCommand(NewUpgradeCmd())
	rootCmd.AddCommand(NewActivateCmd())
	rootCmd.AddCommand(NewReviewCmd())
	rootCmd.AddCommand(NewLogsCmd())
	rootCmd.AddCommand(NewDoctorCmd())
	rootCmd.AddCommand(bootstrapcmd.NewBootstrapCmd())
	rootCmd.AddCommand(configcmd.NewConfigCmd())
	rootCmd.AddCommand(linkcmd.NewLinkCmd())
	rootCmd.AddCommand(provider.NewProviderCmd())
	rootCmd.AddCommand(profile.NewProfileCmd())
	rootCmd.AddCommand(packagecmd.NewPackageCmd())
	rootCmd.AddCommand(shellcmd.NewCmd())

	// Add global JSON flag
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output in JSON format for automation")

	return rootCmd
}
