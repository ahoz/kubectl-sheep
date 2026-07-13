package prompt

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"text/template"

	"github.com/chzyer/readline"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/manifoldco/promptui"
	"github.com/manifoldco/promptui/list"
	"github.com/manifoldco/promptui/screenbuf"
)

const (
	multiSelectHelp = "Use the arrow keys to navigate: \xE2\x86\x93 \xE2\x86\x91 \xE2\x86\x92 \xE2\x86\x90 and / toggles search; space toggles selection; enter confirms"
	hideCursor      = "\033[?25l"
	showCursor      = "\033[?25h"
)

// ChooseMulti shows a list and returns the indices of all selected items.
func ChooseMulti(in io.Reader, out io.Writer, sectionTitle string, choices []Choice) ([]int, error) {
	if len(choices) == 0 {
		return nil, fmt.Errorf("no choices available")
	}
	if canUseInteractiveMenu(in) {
		return chooseMultiInteractive(out, sectionTitle, choices)
	}
	return chooseMultiLine(in, out, sectionTitle, choices)
}

func chooseMultiLine(in io.Reader, out io.Writer, sectionTitle string, choices []Choice) ([]int, error) {
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
	_, _ = fmt.Fprintf(out, "  Choose clusters [1-%d] (comma-separated): ", len(choices))

	line, err := readLine(in)
	if err != nil {
		return nil, fmt.Errorf("read selection: %w", err)
	}

	answer := strings.TrimSpace(line)
	if answer == "" {
		return nil, fmt.Errorf("selection must not be empty")
	}

	parts := strings.Split(answer, ",")
	seen := make(map[int]struct{}, len(parts))
	var indices []int
	for _, part := range parts {
		n, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil || n < 1 || n > len(choices) {
			return nil, fmt.Errorf("invalid selection %q", strings.TrimSpace(part))
		}
		if _, ok := seen[n]; ok {
			continue
		}
		seen[n] = struct{}{}
		indices = append(indices, n-1)
	}
	if len(indices) == 0 {
		return nil, fmt.Errorf("selection must not be empty")
	}
	return indices, nil
}

func chooseMultiInteractive(out io.Writer, sectionTitle string, choices []Choice) ([]int, error) {
	_ = out

	// Separate this menu from prior prompts on the terminal.
	Step(os.Stdout)

	label := strings.TrimSpace(sectionTitle)
	if label == "" {
		label = "Select clusters"
	}

	selected := make([]bool, len(choices))

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

	l, err := list.New(choices, defaultMenuSize)
	if err != nil {
		return nil, err
	}
	l.Searcher = searcher

	stdin := readline.NewCancelableStdin(os.Stdin)
	c := &readline.Config{}
	if err := c.Init(); err != nil {
		return nil, err
	}
	c.Stdin = stdin
	c.HistoryLimit = -1
	c.UniqueEditLine = true

	rl, err := readline.NewEx(c)
	if err != nil {
		return nil, err
	}
	defer rl.Close()

	rl.Write([]byte(hideCursor))
	defer rl.Write([]byte(showCursor))

	sb := screenbuf.New(rl)
	searchMode := false
	search := promptui.NewCursor("", nil, false)
	done := false

	labelTpl, err := template.New("label").Funcs(promptui.FuncMap).Parse(`{{ .Title | cyan }} {{ .Count | faint }}`)
	if err != nil {
		return nil, err
	}

	selectedCount := func() int {
		n := 0
		for _, on := range selected {
			if on {
				n++
			}
		}
		return n
	}

	choiceIndex := func(choice Choice) int {
		for i, ch := range choices {
			if ch.Title == choice.Title && ch.Subtitle == choice.Subtitle {
				return i
			}
		}
		return -1
	}

	render := func() {
		sb.Reset()

		if searchMode {
			sb.Write([]byte(promptui.SearchPrompt + search.Format()))
		} else {
			sb.Write([]byte(multiSelectHelp))
		}

		var labelBuf bytes.Buffer
		_ = labelTpl.Execute(&labelBuf, struct {
			Title string
			Count string
		}{
			Title: label,
			Count: fmt.Sprintf("(%d selected)", selectedCount()),
		})
		sb.Write(labelBuf.Bytes())

		visible, cursor := l.Items()
		last := len(visible) - 1
		activeOrigIdx := l.Index()

		for i, item := range visible {
			choice := item.(Choice)
			origIdx := choiceIndex(choice)
			if i == cursor {
				origIdx = activeOrigIdx
			}

			page := " "
			switch i {
			case 0:
				if l.CanPageUp() {
					page = "↑"
				}
			case last:
				if l.CanPageDown() {
					page = "↓"
				}
			}

			mark := "[ ]"
			if origIdx >= 0 && origIdx < len(selected) && selected[origIdx] {
				mark = "[x]"
			}

			var line string
			if i == cursor {
				line = fmt.Sprintf("%s %s %s %s", page, mark, menuIconActive, choice.Title)
			} else {
				line = fmt.Sprintf("%s %s  %s", page, mark, choice.Title)
			}
			sb.Write([]byte(line))
		}

		if cursor == list.NotFound {
			sb.Write([]byte("No results"))
		} else if activeOrigIdx >= 0 && activeOrigIdx < len(choices) {
			cur := choices[activeOrigIdx]
			sb.Write([]byte("--------- Info ----------"))
			if len(cur.Details) > 0 {
				for _, d := range cur.Details {
					sb.Write([]byte(fmt.Sprintf("%s: %s", d.Label, d.Value)))
				}
			} else {
				sb.Write([]byte(fmt.Sprintf("Name: %s", cur.Title)))
				if cur.Subtitle != "" {
					sb.Write([]byte(fmt.Sprintf("Details: %s", cur.Subtitle)))
				}
			}
		}

		_ = sb.Flush()
	}

	c.SetListener(func(line []rune, pos int, key rune) ([]rune, int, bool) {
		switch {
		case key == promptui.KeyEnter:
			if selectedCount() > 0 {
				done = true
			}
		case key == promptui.KeyNext, key == 'j':
			if !searchMode {
				l.Next()
			}
		case key == promptui.KeyPrev, key == 'k':
			if !searchMode {
				l.Prev()
			}
		case key == promptui.KeyBackward, key == 'h':
			if !searchMode {
				l.PageUp()
			}
		case key == promptui.KeyForward, key == 'l':
			if !searchMode {
				l.PageDown()
			}
		case key == '/':
			if l.Searcher == nil {
				break
			}
			if searchMode {
				searchMode = false
				search.Replace("")
				l.CancelSearch()
			} else {
				searchMode = true
			}
		case key == promptui.KeyBackspace:
			if searchMode {
				search.Backspace()
				if len(search.Get()) > 0 {
					l.Search(string(search.Get()))
				} else {
					l.CancelSearch()
				}
			}
		case key == ' ':
			if !searchMode {
				if idx := l.Index(); idx >= 0 && idx < len(selected) {
					selected[idx] = !selected[idx]
				}
			}
		default:
			if searchMode && key >= 32 {
				search.Update(string(line))
				l.Search(string(search.Get()))
			}
		}

		render()
		return nil, 0, true
	})

	var runErr error
	for {
		_, runErr = rl.Readline()
		if runErr != nil {
			switch {
			case runErr == readline.ErrInterrupt:
				runErr = promptui.ErrInterrupt
			case runErr.Error() == "Interrupt":
				runErr = promptui.ErrInterrupt
			}
			break
		}
		if done {
			break
		}
	}

	if runErr != nil {
		sb.Reset()
		sb.Write([]byte(""))
		_ = sb.Flush()
		return nil, runErr
	}

	var indices []int
	for i, on := range selected {
		if on {
			indices = append(indices, i)
		}
	}
	return indices, nil
}
