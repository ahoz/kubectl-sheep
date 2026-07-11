package cmd

import (
	"errors"
	"fmt"
	"testing"

	"github.com/ahoz/kubectl-sheep/internal/rancher"
)

func TestHandleRancherErrorTokenInvalid(t *testing.T) {
	err := handleRancherError("prod", rancher.ErrTokenInvalid)
	if err == nil {
		t.Fatal("expected error")
	}
	if got := err.Error(); got != `rancher token for rancher-instance "prod" is invalid or expired; run: kubectl sheep rancher-instance update-token prod` {
		t.Fatalf("unexpected message: %s", got)
	}
}

func TestHandleRancherErrorPassthrough(t *testing.T) {
	orig := fmt.Errorf("network down")
	if got := handleRancherError("prod", orig); !errors.Is(got, orig) {
		t.Fatalf("expected passthrough, got %v", got)
	}
}
