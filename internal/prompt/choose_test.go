package prompt

import (
	"bytes"
	"strings"
	"testing"
)

func TestChooseByNumber(t *testing.T) {
	choices := []Choice{
		{Title: "main", Subtitle: "https://127.0.0.1/"},
	}
	idx, free, err := Choose(strings.NewReader("1\n"), &bytes.Buffer{}, "Rancher instance", choices)
	if err != nil {
		t.Fatalf("Choose: %v", err)
	}
	if idx != 0 || free != "" {
		t.Fatalf("got index=%d free=%q", idx, free)
	}
}

func TestChooseFreeText(t *testing.T) {
	choices := []Choice{{Title: "main"}, {Title: "prod"}}
	idx, free, err := Choose(strings.NewReader("staging\n"), &bytes.Buffer{}, "Rancher instance", choices)
	if err != nil {
		t.Fatalf("Choose: %v", err)
	}
	if idx != -1 || free != "staging" {
		t.Fatalf("got index=%d free=%q", idx, free)
	}
}
