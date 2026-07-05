package credentials

import (
	"fmt"

	"github.com/ahoz/kubectl-sheep/internal/config"
)

// NewStore returns the credential store for the given storage mode.
func NewStore(storage string) (Store, error) {
	switch storage {
	case config.StoragePlaintext:
		return NewPlaintextStore()
	case config.StorageEncrypted:
		return NewEncryptedStore()
	default:
		return nil, fmt.Errorf("unknown storage mode %q", storage)
	}
}

// MigrateStorage moves a token from one backend to another.
func MigrateStorage(instance, from, to string) error {
	if from == to {
		return fmt.Errorf("instance %q is already using %q storage", instance, to)
	}
	if err := config.ValidateStorage(to); err != nil {
		return err
	}

	src, err := NewStore(from)
	if err != nil {
		return err
	}
	dst, err := NewStore(to)
	if err != nil {
		return err
	}

	token, err := src.Get(instance)
	if err != nil {
		return fmt.Errorf("read token from %s storage: %w", from, err)
	}
	if err := dst.Set(instance, token); err != nil {
		return fmt.Errorf("write token to %s storage: %w", to, err)
	}
	if err := src.Delete(instance); err != nil {
		return fmt.Errorf("delete token from %s storage: %w", from, err)
	}
	return nil
}
