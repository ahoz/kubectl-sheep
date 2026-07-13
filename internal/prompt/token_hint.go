package prompt

import "io"

// PrintTokenCreateHint prints the Rancher UI URL for creating an API key.
func PrintTokenCreateHint(out io.Writer, tokenPageURL string) {
	if tokenPageURL == "" {
		return
	}
	Step(out)
	Note(out, "Create a Rancher API key (copy the Bearer Token) at:")
	Note(out, tokenPageURL)
}
