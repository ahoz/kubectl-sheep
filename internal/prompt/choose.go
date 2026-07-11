package prompt

import (
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Choice is one row in a numbered interactive list.
type Choice struct {
	Title    string
	Subtitle string
}

// Choose shows a numbered list with optional subtitles and returns the chosen
// index, or -1 if the user entered a free-form value that did not match a number.
func Choose(in io.Reader, out io.Writer, sectionTitle string, choices []Choice) (index int, freeText string, err error) {
	if sectionTitle != "" {
		Section(out, sectionTitle)
	}
	for i, choice := range choices {
		_, _ = fmt.Fprintf(out, "  %d  %s\n", i+1, choice.Title)
		if strings.TrimSpace(choice.Subtitle) != "" {
			Note(out, choice.Subtitle)
		}
	}
	Blank(out)
	_, _ = fmt.Fprintf(out, "  Choose [1-%d] (or type a value): ", len(choices))

	line, err := readLine(in)
	if err != nil {
		return -1, "", fmt.Errorf("read selection: %w", err)
	}

	answer := strings.TrimSpace(line)
	if answer == "" {
		return -1, "", fmt.Errorf("selection must not be empty")
	}

	if n, err := strconv.Atoi(answer); err == nil {
		if n < 1 || n > len(choices) {
			return -1, "", fmt.Errorf("selection %d is out of range", n)
		}
		return n - 1, "", nil
	}

	return -1, answer, nil
}
