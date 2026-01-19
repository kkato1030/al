package cmd

import (
	"fmt"
	"strings"

	configcmd "github.com/kkato1030/al/cmd/config"
	packagecmd "github.com/kkato1030/al/cmd/package"
	"github.com/kkato1030/al/cmd/profile"
	"github.com/kkato1030/al/cmd/provider"
	"github.com/kkato1030/al/internal/config"
	"github.com/spf13/cobra"
)

var version = "0.1.0"

// SetVersion sets the version string
func SetVersion(v string) {
	version = v
}

// buildAliasSection builds the alias section string for help
func buildAliasSection() string {
	aliases := config.GetDefaultAliases()
	if len(aliases) == 0 {
		return ""
	}
	
	var sb strings.Builder
	sb.WriteString("\n\nDefault Aliases:")
	aliasNames := []string{"add", "remove", "list", "promote", "register"}
	for _, name := range aliasNames {
		if cmdStr, exists := aliases[name]; exists {
			sb.WriteString(fmt.Sprintf("\n  %-10s %s", name, cmdStr))
		}
	}
	return sb.String()
}

// NewRootCmd creates the root command
func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "al",
		Short: "Mac Management Tools",
		Long:  "al - Mac Management Tools",
	}

	// Build custom help template with aliases after Available Commands
	aliasSection := buildAliasSection()
	helpTemplate := `{{with (or .Long .Short)}}{{. | trimTrailingWhitespaces}}

{{end}}{{if or .Runnable .HasSubCommands}}
Usage:
{{if .Runnable}}  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}  {{.UseLine}} [command]{{end}}{{end}}{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}` + aliasSection + `{{if .HasAvailableLocalFlags}}

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
	rootCmd.AddCommand(NewUpdateCmd())
	rootCmd.AddCommand(NewUpgradeCmd())
	rootCmd.AddCommand(configcmd.NewConfigCmd())
	rootCmd.AddCommand(provider.NewProviderCmd())
	rootCmd.AddCommand(profile.NewProfileCmd())
	rootCmd.AddCommand(packagecmd.NewPackageCmd())

	return rootCmd
}
