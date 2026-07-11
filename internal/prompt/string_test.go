package prompt

import (
	"bytes"
	"strings"
	"testing"
)

func TestReadStringUsesDefault(t *testing.T) {
	got, err := ReadString(strings.NewReader("\n"), &bytes.Buffer{}, "Name", "prod")
	if err != nil {
		t.Fatalf("ReadString: %v", err)
	}
	if got != "prod" {
		t.Fatalf("got %q, want prod", got)
	}
}

func TestReadStringUsesInput(t *testing.T) {
	got, err := ReadString(strings.NewReader("dev\n"), &bytes.Buffer{}, "Name", "prod")
	if err != nil {
		t.Fatalf("ReadString: %v", err)
	}
	if got != "dev" {
		t.Fatalf("got %q, want dev", got)
	}
}

func TestSelectByNumber(t *testing.T) {
	idx, free, err := Select(strings.NewReader("2\n"), &bytes.Buffer{}, "Pick", []string{"a", "b"})
	if err != nil {
		t.Fatalf("Select: %v", err)
	}
	if idx != 1 || free != "" {
		t.Fatalf("got index=%d free=%q", idx, free)
	}
}

func TestSelectFreeText(t *testing.T) {
	idx, free, err := Select(strings.NewReader("my-cluster\n"), &bytes.Buffer{}, "Pick", []string{"a", "b"})
	if err != nil {
		t.Fatalf("Select: %v", err)
	}
	if idx != -1 || free != "my-cluster" {
		t.Fatalf("got index=%d free=%q", idx, free)
	}
}
