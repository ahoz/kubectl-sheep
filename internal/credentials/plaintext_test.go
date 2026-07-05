package credentials

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPlaintextStoreRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "credentials.plain.yaml")
	store := NewPlaintextStoreAt(path)

	if err := store.Set("prod", "secret-token"); err != nil {
		t.Fatalf("Set: %v", err)
	}

	token, err := store.Get("prod")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if token != "secret-token" {
		t.Fatalf("got token %q, want %q", token, "secret-token")
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("expected mode 0600, got %o", info.Mode().Perm())
	}

	if err := store.Delete("prod"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := store.Get("prod"); err == nil {
		t.Fatal("expected error after delete")
	}
}

func TestPlaintextStoreMultipleInstances(t *testing.T) {
	path := filepath.Join(t.TempDir(), "credentials.plain.yaml")
	store := NewPlaintextStoreAt(path)

	_ = store.Set("a", "token-a")
	_ = store.Set("b", "token-b")

	got, err := store.Get("b")
	if err != nil || got != "token-b" {
		t.Fatalf("Get(b): %q, %v", got, err)
	}

	_ = store.Delete("a")
	if _, err := store.Get("a"); err == nil {
		t.Fatal("expected missing token error")
	}
	got, _ = store.Get("b")
	if got != "token-b" {
		t.Fatalf("b should remain, got %q", got)
	}
}
