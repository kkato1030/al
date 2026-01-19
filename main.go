package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/kkato1030/al/cmd"
	"github.com/kkato1030/al/internal/config"
)

func main() {
	// Resolve aliases before executing
	resolvedArgs, err := resolveAliases(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving aliases: %v\n", err)
		os.Exit(1)
	}

	// Create new args with resolved alias
	newArgs := []string{os.Args[0]}
	newArgs = append(newArgs, resolvedArgs...)

	// Update os.Args for cobra
	os.Args = newArgs

	rootCmd := cmd.NewRootCmd()

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// resolveAliases resolves command aliases
func resolveAliases(args []string) ([]string, error) {
	if len(args) == 0 {
		return args, nil
	}

	// Get default aliases
	aliases := config.GetDefaultAliases()

	// Check if first argument is an alias
	aliasName := args[0]
	aliasCommand, exists := aliases[aliasName]
	if !exists {
		return args, nil
	}

	// Resolve special variables in alias command
	originalArgs := args[1:]
	resolvedCommand, err := resolveAliasVariables(aliasCommand, originalArgs)
	if err != nil {
		return nil, fmt.Errorf("error resolving alias variables: %w", err)
	}

	// Parse the resolved command into arguments
	resolvedArgs := parseCommand(resolvedCommand)

	// Insert original arguments at appropriate positions
	// Replace {args} placeholder if present, otherwise append at the end
	resolvedArgs = insertArguments(resolvedArgs, originalArgs)

	return resolvedArgs, nil
}

// resolveAliasVariables resolves special variables like package.promote_to in alias commands
func resolveAliasVariables(aliasCommand string, originalArgs []string) (string, error) {
	result := aliasCommand

	// Resolve package.promote_to
	if strings.Contains(result, "package.promote_to") {
		promoteTo, err := resolvePromoteTo(originalArgs)
		if err != nil {
			return "", fmt.Errorf("error resolving package.promote_to: %w", err)
		}
		if promoteTo == "" {
			return "", fmt.Errorf("package.promote_to could not be resolved. Make sure the package exists and its profile has promote_to set")
		}
		result = strings.ReplaceAll(result, "package.promote_to", promoteTo)
	}

	return result, nil
}

// resolvePromoteTo resolves the promote_to value for a package
func resolvePromoteTo(args []string) (string, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("package name is required")
	}

	packageName := args[0]

	// Load packages config to find the package
	packagesConfig, err := config.LoadPackagesConfig()
	if err != nil {
		return "", err
	}

	// Find the package
	var foundPackage *config.PackageConfig
	for i := range packagesConfig.Packages {
		if packagesConfig.Packages[i].Name == packageName {
			foundPackage = &packagesConfig.Packages[i]
			break
		}
	}

	if foundPackage == nil {
		// Try to use default profile's promote_to
		appConfig, err := config.LoadAppConfig()
		if err != nil {
			return "", err
		}

		if appConfig.DefaultProfile != "" {
			profile, err := config.GetProfile(appConfig.DefaultProfile)
			if err != nil {
				return "", err
			}
			if profile != nil && profile.PromoteTo != "" {
				return profile.PromoteTo, nil
			}
		}

		return "", fmt.Errorf("package '%s' not found", packageName)
	}

	// Get the profile's promote_to
	profile, err := config.GetProfile(foundPackage.Profile)
	if err != nil {
		return "", err
	}

	if profile == nil {
		return "", fmt.Errorf("profile '%s' not found", foundPackage.Profile)
	}

	if profile.PromoteTo == "" {
		return "", fmt.Errorf("profile '%s' does not have promote_to set", foundPackage.Profile)
	}

	return profile.PromoteTo, nil
}

// insertArguments inserts original arguments into resolved command
// Replaces {args} placeholder if present, otherwise appends at the end
func insertArguments(resolvedArgs, originalArgs []string) []string {
	if len(originalArgs) == 0 {
		return resolvedArgs
	}

	// Check if {args} placeholder exists
	for i, arg := range resolvedArgs {
		if arg == "{args}" {
			// Replace {args} with original arguments
			newArgs := make([]string, 0, len(resolvedArgs)-1+len(originalArgs))
			newArgs = append(newArgs, resolvedArgs[:i]...)
			newArgs = append(newArgs, originalArgs...)
			newArgs = append(newArgs, resolvedArgs[i+1:]...)
			return newArgs
		}
	}

	// No placeholder found, append at the end
	return append(resolvedArgs, originalArgs...)
}

// parseCommand parses a command string into arguments
func parseCommand(command string) []string {
	if command == "" {
		return []string{}
	}

	// Simple parsing: split by spaces, but preserve quoted strings
	var args []string
	var current strings.Builder
	inQuotes := false
	quoteChar := byte(0)

	for i := 0; i < len(command); i++ {
		char := command[i]

		if char == '"' || char == '\'' {
			if inQuotes && char == quoteChar {
				// End of quoted string
				inQuotes = false
				quoteChar = 0
			} else if !inQuotes {
				// Start of quoted string
				inQuotes = true
				quoteChar = char
			} else {
				current.WriteByte(char)
			}
		} else if char == ' ' && !inQuotes {
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
		} else {
			current.WriteByte(char)
		}
	}

	if current.Len() > 0 {
		args = append(args, current.String())
	}

	return args
}
