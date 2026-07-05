package prompt

import (
	"fmt"
	"io"
)

// PrintTokenCreateHint prints the Rancher UI URL for creating an API key.
func PrintTokenCreateHint(out io.Writer, tokenPageURL string) {
	if tokenPageURL == "" {
		return
	}
	_, _ = fmt.Fprintf(out, "Create a Rancher API key (copy the Bearer Token) at:\n  %s\n\n", tokenPageURL)
}
