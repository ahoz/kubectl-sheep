package prompt

import (
	"fmt"
	"io"
	"strconv"
	"strings"
)

// ReadString prompts for a line of input. An empty answer returns defaultValue.
func ReadString(in io.Reader, out io.Writer, label, defaultValue string) (string, error) {
	if defaultValue != "" {
		_, _ = fmt.Fprintf(out, "%s [%s]: ", label, defaultValue)
	} else {
		_, _ = fmt.Fprintf(out, "%s: ", label)
	}

	line, err := readLine(in)
	if err != nil {
		return "", fmt.Errorf("read input: %w", err)
	}

	answer := strings.TrimSpace(line)
	if answer == "" {
		return defaultValue, nil
	}
	return answer, nil
}

// Select shows a numbered list and returns the chosen index, or -1 if the user
// entered a free-form value that did not match a number.
func Select(in io.Reader, out io.Writer, title string, options []string) (index int, freeText string, err error) {
	if title != "" {
		_, _ = fmt.Fprintln(out, title)
	}
	for i, opt := range options {
		_, _ = fmt.Fprintf(out, "  %d) %s\n", i+1, opt)
	}
	_, _ = fmt.Fprint(out, "Select [1-", len(options), "] (or type a value): ")

	line, err := readLine(in)
	if err != nil {
		return -1, "", fmt.Errorf("read selection: %w", err)
	}

	answer := strings.TrimSpace(line)
	if answer == "" {
		return -1, "", fmt.Errorf("selection must not be empty")
	}

	if n, err := strconv.Atoi(answer); err == nil {
		if n < 1 || n > len(options) {
			return -1, "", fmt.Errorf("selection %d is out of range", n)
		}
		return n - 1, "", nil
	}

	return -1, answer, nil
}
