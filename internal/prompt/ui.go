package prompt

import (
	"fmt"
	"io"
)

// Blank prints a single empty line.
func Blank(out io.Writer) {
	_, _ = fmt.Fprintln(out)
}

// Intro prints a short heading for an interactive command flow.
func Intro(out io.Writer, title string) {
	_, _ = fmt.Fprintf(out, "\n%s\n", title)
}

// Step separates sequential interactive prompts.
func Step(out io.Writer) {
	Blank(out)
}

// Note prints an indented secondary line.
func Note(out io.Writer, text string) {
	_, _ = fmt.Fprintf(out, "  %s\n", text)
}

// Success prints a positive completion line.
func Success(out io.Writer, message string) {
	Blank(out)
	_, _ = fmt.Fprintf(out, "✓ %s\n", message)
}

// Section prints a titled divider for non-interactive prompt steps.
func Section(out io.Writer, title string) {
	Step(out)
	_, _ = fmt.Fprintf(out, "── %s ──\n\n", title)
}
