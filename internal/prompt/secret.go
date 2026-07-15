package prompt

import (
	"fmt"
	"io"
	"os"

	"golang.org/x/term"
)

// ReadSecret prompts for a secret value with hidden input.
// Input length is not echoed to the terminal.
// An optional hint is printed as a Note above the prompt line.
func ReadSecret(in *os.File, out io.Writer, label string, hint ...string) (string, error) {
	Step(out)
	if len(hint) > 0 && hint[0] != "" {
		Note(out, hint[0])
	}
	_, _ = fmt.Fprintf(out, "  %s: ", label)
	bytes, err := term.ReadPassword(int(in.Fd()))
	_, _ = fmt.Fprintln(out)
	if err != nil {
		return "", fmt.Errorf("read secret input: %w", err)
	}
	return string(bytes), nil
}
