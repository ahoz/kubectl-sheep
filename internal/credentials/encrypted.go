package credentials

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/99designs/keyring"
	"github.com/ahoz/kubectl-sheep/internal/config"
	"github.com/ahoz/kubectl-sheep/internal/prompt"
)

const (
	keyringService = "kubectl-sheep"
	tokenKeyPrefix = "rancher-token:"
)

// EncryptedStore stores tokens using keyring FileBackend with a passphrase.
type EncryptedStore struct {
	ring keyring.Keyring
}

var (
	encryptedStoreMu sync.Mutex
	encryptedStore   *EncryptedStore
)

// NewEncryptedStore opens the encrypted keyring at the default keys directory.
// The opened keyring is cached for the process so the passphrase is requested once.
func NewEncryptedStore() (*EncryptedStore, error) {
	encryptedStoreMu.Lock()
	defer encryptedStoreMu.Unlock()
	if encryptedStore != nil {
		return encryptedStore, nil
	}

	dir, err := config.ConfigDir()
	if err != nil {
		return nil, err
	}
	store, err := NewEncryptedStoreAt(filepath.Join(dir, "keys"))
	if err != nil {
		return nil, err
	}
	encryptedStore = store
	return store, nil
}

// NewEncryptedStoreAt opens an encrypted keyring at dir (for tests).
func NewEncryptedStoreAt(dir string) (*EncryptedStore, error) {
	return newEncryptedStoreAt(dir, passphrasePrompt)
}

func passphrasePrompt(_ string) (string, error) {
	if !prompt.IsTerminal(os.Stdin) {
		return keyring.TerminalPrompt("Enter passphrase for encrypted token storage")
	}
	prompt.Section(os.Stdout, "Encrypted token storage")
	return prompt.ReadSecret(os.Stdin, os.Stdout, "Passphrase")
}

// NewEncryptedStoreAtWithPassword opens an encrypted keyring with a fixed passphrase (for tests).
func NewEncryptedStoreAtWithPassword(dir, passphrase string) (*EncryptedStore, error) {
	return newEncryptedStoreAt(dir, func(_ string) (string, error) {
		return passphrase, nil
	})
}

func newEncryptedStoreAt(dir string, passwordFunc func(string) (string, error)) (*EncryptedStore, error) {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("create keys directory: %w", err)
	}

	ring, err := keyring.Open(keyring.Config{
		AllowedBackends: []keyring.BackendType{keyring.FileBackend},
		ServiceName:     keyringService,
		FileDir:         dir,
		FilePasswordFunc: func(prompt string) (string, error) {
			return passwordFunc(prompt)
		},
	})
	if err != nil {
		return nil, fmt.Errorf("open encrypted keyring: %w", err)
	}
	return &EncryptedStore{ring: ring}, nil
}

func tokenKey(instance string) string {
	return tokenKeyPrefix + instance
}

// Get returns the token for instance.
func (s *EncryptedStore) Get(instance string) (string, error) {
	item, err := s.ring.Get(tokenKey(instance))
	if err != nil {
		return "", fmt.Errorf("get encrypted token for instance %q: %w", instance, err)
	}
	return string(item.Data), nil
}

// Set stores the token for instance.
func (s *EncryptedStore) Set(instance, token string) error {
	if err := s.ring.Set(keyring.Item{
		Key:  tokenKey(instance),
		Data: []byte(token),
	}); err != nil {
		return fmt.Errorf("set encrypted token for instance %q: %w", instance, err)
	}
	return nil
}

// Delete removes the token for instance.
func (s *EncryptedStore) Delete(instance string) error {
	if err := s.ring.Remove(tokenKey(instance)); err != nil {
		return fmt.Errorf("delete encrypted token for instance %q: %w", instance, err)
	}
	return nil
}
