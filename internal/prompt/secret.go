package prompt

import (
	"fmt"
	"io"
	"os"

	"golang.org/x/term"
)

// ReadSecret prompts for a secret value with hidden input.
func ReadSecret(in *os.File, out io.Writer, label string) (string, error) {
	_, _ = fmt.Fprintf(out, "  %s: ", label)
	bytes, err := term.ReadPassword(int(in.Fd()))
	_, _ = fmt.Fprintln(out)
	if err != nil {
		return "", fmt.Errorf("read secret input: %w", err)
	}
	return string(bytes), nil
}
