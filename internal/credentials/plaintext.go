package credentials

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/ahoz/kubectl-sheep/internal/config"
	"gopkg.in/yaml.v3"
)

const credentialsFileName = "credentials.plain.yaml"

// PlaintextStore stores tokens in a YAML file with mode 0600.
type PlaintextStore struct {
	path string
	mu   sync.Mutex
}

// NewPlaintextStore returns a PlaintextStore using the default config path.
func NewPlaintextStore() (*PlaintextStore, error) {
	dir, err := config.ConfigDir()
	if err != nil {
		return nil, err
	}
	return &PlaintextStore{path: filepath.Join(dir, credentialsFileName)}, nil
}

// NewPlaintextStoreAt creates a PlaintextStore at an explicit path (for tests).
func NewPlaintextStoreAt(path string) *PlaintextStore {
	return &PlaintextStore{path: path}
}

type plaintextData struct {
	Tokens map[string]string `yaml:"tokens"`
}

func (s *PlaintextStore) load() (plaintextData, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return plaintextData{Tokens: map[string]string{}}, nil
		}
		return plaintextData{}, fmt.Errorf("read plaintext credentials: %w", err)
	}

	var stored plaintextData
	if err := yaml.Unmarshal(data, &stored); err != nil {
		return plaintextData{}, fmt.Errorf("parse plaintext credentials: %w", err)
	}
	if stored.Tokens == nil {
		stored.Tokens = map[string]string{}
	}
	return stored, nil
}

func (s *PlaintextStore) save(tokens map[string]string) error {
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create credentials directory: %w", err)
	}

	data, err := yaml.Marshal(plaintextData{Tokens: tokens})
	if err != nil {
		return fmt.Errorf("marshal plaintext credentials: %w", err)
	}
	if err := os.WriteFile(s.path, data, 0o600); err != nil {
		return fmt.Errorf("write plaintext credentials: %w", err)
	}
	return nil
}

// Get returns the token for instance.
func (s *PlaintextStore) Get(instance string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	stored, err := s.load()
	if err != nil {
		return "", err
	}
	token, ok := stored.Tokens[instance]
	if !ok {
		return "", fmt.Errorf("no plaintext token for instance %q", instance)
	}
	return token, nil
}

// Set stores the token for instance.
func (s *PlaintextStore) Set(instance, token string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	stored, err := s.load()
	if err != nil {
		return err
	}
	stored.Tokens[instance] = token
	return s.save(stored.Tokens)
}

// Delete removes the token for instance.
func (s *PlaintextStore) Delete(instance string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	stored, err := s.load()
	if err != nil {
		return err
	}
	delete(stored.Tokens, instance)
	return s.save(stored.Tokens)
}
