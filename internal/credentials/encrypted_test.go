package credentials

import (
	"errors"
	"strings"
	"testing"
)

func TestEncryptedStoreRoundTrip(t *testing.T) {
	dir := t.TempDir()
	store, err := NewEncryptedStoreAtWithPassword(dir, "test-pass")
	if err != nil {
		t.Fatalf("NewEncryptedStoreAt: %v", err)
	}

	if err := store.Set("prod", "encrypted-secret"); err != nil {
		t.Fatalf("Set: %v", err)
	}

	token, err := store.Get("prod")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if token != "encrypted-secret" {
		t.Fatalf("got token %q, want %q", token, "encrypted-secret")
	}

	if err := store.Delete("prod"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := store.Get("prod"); err == nil {
		t.Fatal("expected error after delete")
	}
}

func TestMigrateStoragePlainToEncrypted(t *testing.T) {
	tmp := t.TempDir()
	plainPath := tmp + "/credentials.plain.yaml"
	keysDir := tmp + "/keys"

	plain := NewPlaintextStoreAt(plainPath)
	if err := plain.Set("dev", "migrate-me"); err != nil {
		t.Fatalf("Set: %v", err)
	}

	// Migrate using explicit stores to avoid default paths.
	src := plain
	dst, err := NewEncryptedStoreAtWithPassword(keysDir, "test-pass")
	if err != nil {
		t.Fatalf("NewEncryptedStoreAt: %v", err)
	}

	token, err := src.Get("dev")
	if err != nil {
		t.Fatalf("src Get: %v", err)
	}
	if err := dst.Set("dev", token); err != nil {
		t.Fatalf("dst Set: %v", err)
	}
	if err := src.Delete("dev"); err != nil {
		t.Fatalf("src Delete: %v", err)
	}

	got, err := dst.Get("dev")
	if err != nil || got != "migrate-me" {
		t.Fatalf("dst Get: %q, %v", got, err)
	}
	if _, err := src.Get("dev"); err == nil {
		t.Fatal("plaintext token should be deleted")
	}
}

func TestEncryptedStoreWrongPassphrase(t *testing.T) {
	dir := t.TempDir()

	store, err := NewEncryptedStoreAtWithPassword(dir, "correct-pass")
	if err != nil {
		t.Fatalf("NewEncryptedStoreAtWithPassword: %v", err)
	}
	if err := store.Set("prod", "secret"); err != nil {
		t.Fatalf("Set: %v", err)
	}

	encryptedStoreMu.Lock()
	encryptedStore = store
	encryptedStoreMu.Unlock()

	wrong, err := newEncryptedStoreAt(dir, func(_ string) (string, error) {
		return "wrong-pass", nil
	})
	if err != nil {
		t.Fatalf("newEncryptedStoreAt: %v", err)
	}

	_, err = wrong.Get("prod")
	if !errors.Is(err, ErrWrongPassphrase) {
		t.Fatalf("Get: got %v, want ErrWrongPassphrase", err)
	}
	if !strings.Contains(err.Error(), `rancher-instance "prod"`) {
		t.Fatalf("Get: expected instance in error, got %v", err)
	}

	encryptedStoreMu.Lock()
	cached := encryptedStore
	encryptedStoreMu.Unlock()
	if cached != nil {
		t.Fatal("expected encrypted store cache to be cleared after wrong passphrase")
	}
}
