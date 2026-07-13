package prompt

import (
	"io"
	"os"
	"strings"

	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/manifoldco/promptui"
)

const defaultMenuSize = 10

// Standard Unicode sheep used as the interactive menu pointer.
const (
	menuIconActive   = "\U0001F411" // 🐑
	menuIconSelected = "\U0001F411" // 🐑
)

func canUseInteractiveMenu(in io.Reader) bool {
	f, ok := in.(*os.File)
	return ok && IsTerminal(f)
}

func chooseInteractive(out io.Writer, sectionTitle string, choices []Choice) (index int, freeText string, err error) {
	_ = out

	label := strings.TrimSpace(sectionTitle)
	if label == "" {
		label = "Select"
	}

	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   menuIconActive + " {{ .Title | red }}",
		Inactive: " {{ .Title | cyan }}",
		Selected: menuIconSelected + " {{ .Title | green }}",
		Details: `--------- Info ----------
{{range .Details}}{{ .Label | printf "%s:" | faint }} {{ .Value }}
{{else}}{{ "Name:" | faint }} {{ .Title }}
{{if .Subtitle }}{{ "Details:" | faint }} {{ .Subtitle }}
{{end}}{{end}}`,
	}

	searcher := func(input string, index int) bool {
		choice := choices[index]
		title := strings.ReplaceAll(strings.ToLower(choice.Title), " ", "")
		subtitle := strings.ReplaceAll(strings.ToLower(choice.Subtitle), " ", "")
		input = strings.ReplaceAll(strings.ToLower(input), " ", "")
		if input == "" {
			return true
		}
		return fuzzy.Match(input, title) || (subtitle != "" && fuzzy.Match(input, subtitle))
	}

	prompt := promptui.Select{
		Label:     label,
		Items:     choices,
		Templates: templates,
		Size:      defaultMenuSize,
		Searcher:  searcher,
	}

	i, _, err := prompt.Run()
	if err != nil {
		return -1, "", err
	}
	return i, "", nil
}

func selectInteractive(out io.Writer, title string, options []string) (index int, freeText string, err error) {
	_ = out

	label := strings.TrimSpace(title)
	if label == "" {
		label = "Select"
	}

	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   menuIconActive + " {{ . | red }}",
		Inactive: " {{ . | cyan }}",
		Selected: menuIconSelected + " {{ . | green }}",
	}

	searcher := func(input string, index int) bool {
		name := strings.ReplaceAll(strings.ToLower(options[index]), " ", "")
		input = strings.ReplaceAll(strings.ToLower(input), " ", "")
		if input == "" {
			return true
		}
		return fuzzy.Match(input, name)
	}

	prompt := promptui.Select{
		Label:     label,
		Items:     options,
		Templates: templates,
		Size:      defaultMenuSize,
		Searcher:  searcher,
	}

	i, _, err := prompt.Run()
	if err != nil {
		return -1, "", err
	}
	return i, "", nil
}
