package prompt

import (
	"io"
)

func confirmInteractive(out io.Writer, question string, defaultYes bool) (bool, error) {
	_ = defaultYes
	idx, _, err := selectInteractive(out, question, []string{"Yes", "No"})
	if err != nil {
		return false, err
	}
	return idx == 0, nil
}
