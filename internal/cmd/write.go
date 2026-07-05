package cmd

import (
	"fmt"
	"io"
)

// fprint and fprintln ignore write errors (broken pipe, closed stdout, etc.).
func fprint(w io.Writer, format string, a ...any) {
	_, _ = fmt.Fprintf(w, format, a...)
}

func fprintln(w io.Writer, a ...any) {
	_, _ = fmt.Fprintln(w, a...)
}
