package prompt

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// Confirm prints the prompt to w (typically os.Stderr), reads a line from stdin,
// and returns true if the user answered "y" or "yes" (case-insensitive).
// Empty or any other input returns false.
func Confirm(w io.Writer, prompt string) (bool, error) {
	fmt.Fprint(w, prompt)
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return false, err
		}
		return false, nil
	}
	s := strings.TrimSpace(strings.ToLower(scanner.Text()))
	return s == "y" || s == "yes", nil
}
