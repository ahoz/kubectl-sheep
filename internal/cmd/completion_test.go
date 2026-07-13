package cmd

import (
	"testing"
)

func TestFilterCompletions(t *testing.T) {
	got := filterCompletions([]string{"paas", "prod", "dev"}, "p")
	if len(got) != 2 || got[0] != "paas" || got[1] != "prod" {
		t.Fatalf("filterCompletions: %v", got)
	}
}

func TestCompleteRancherInstancesEmptyConfig(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	names, err := listRancherInstanceNames()
	if err != nil {
		t.Fatalf("listRancherInstanceNames: %v", err)
	}
	if len(names) != 0 {
		t.Fatalf("expected no instances, got %v", names)
	}
}
