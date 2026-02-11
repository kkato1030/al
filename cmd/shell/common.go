package shell

import (
	"os"
	"strings"
)

// shellExtFromEnv returns the snippet file extension (.zsh or .bash) based on SHELL.
func shellExtFromEnv() string {
	s := os.Getenv("SHELL")
	if s == "" {
		return ".zsh"
	}
	if strings.Contains(s, "zsh") {
		return ".zsh"
	}
	if strings.Contains(s, "bash") {
		return ".bash"
	}
	return ".zsh"
}
