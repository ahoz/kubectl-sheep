package prompt

import (
	"bytes"
	"strings"
	"testing"
)

func TestConfirmTwice(t *testing.T) {
	var out bytes.Buffer
	in := strings.NewReader("y\ny\n")
	a, err := Confirm(in, &out, "q1", false)
	if err != nil || !a {
		t.Fatalf("first: a=%v err=%v", a, err)
	}
	b, err := Confirm(in, &out, "q2", false)
	if err != nil || !b {
		t.Fatalf("second: b=%v err=%v out=%q", b, err, out.String())
	}
}

func TestConfirm(t *testing.T) {
	tests := []struct {
		input      string
		defaultYes bool
		want       bool
	}{
		{"y\n", false, true},
		{"yes\n", false, true},
		{"n\n", false, false},
		{"\n", false, false},
		{"\n", true, true},
		{"n\n", true, false},
	}

	for _, tt := range tests {
		got, err := Confirm(strings.NewReader(tt.input), &bytes.Buffer{}, "question", tt.defaultYes)
		if err != nil {
			t.Fatalf("input %q: %v", tt.input, err)
		}
		if got != tt.want {
			t.Fatalf("input %q defaultYes=%v: got %v, want %v", tt.input, tt.defaultYes, got, tt.want)
		}
	}
}
