package prompt

import (
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

// IsTerminal reports whether f is connected to an interactive terminal.
func IsTerminal(f *os.File) bool {
	return term.IsTerminal(int(f.Fd()))
}

// Confirm asks a yes/no question. defaultYes controls the answer on empty input.
func Confirm(in io.Reader, out io.Writer, question string, defaultYes bool) (bool, error) {
	defaultHint := "y/N"
	if defaultYes {
		defaultHint = "Y/n"
	}
	_, _ = fmt.Fprintf(out, "%s [%s]: ", question, defaultHint)

	line, err := readLine(in)
	if err != nil {
		return false, fmt.Errorf("read confirmation: %w", err)
	}

	answer := strings.TrimSpace(strings.ToLower(line))
	if answer == "" {
		return defaultYes, nil
	}
	switch answer {
	case "y", "yes":
		return true, nil
	case "n", "no":
		return false, nil
	default:
		return defaultYes, nil
	}
}

func readLine(in io.Reader) (string, error) {
	var line []byte
	buf := make([]byte, 1)
	for {
		n, err := in.Read(buf)
		if n == 1 {
			if buf[0] == '\n' {
				break
			}
			line = append(line, buf[0])
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
	}
	return string(line), nil
}
