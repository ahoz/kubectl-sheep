package instance

import (
	"errors"
	"fmt"

	"github.com/ahoz/kubectl-sheep/internal/config"
	"github.com/ahoz/kubectl-sheep/internal/credentials"
	"github.com/ahoz/kubectl-sheep/internal/rancher"
)

// RancherClient loads instance config and returns an authenticated Rancher client.
func RancherClient(name string) (*config.Instance, *rancher.Client, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, nil, err
	}

	inst, err := cfg.Find(name)
	if err != nil {
		return nil, nil, err
	}

	store, err := credentials.NewStore(inst.Storage)
	if err != nil {
		return nil, nil, err
	}

	token, err := store.Get(name)
	if err != nil {
		if errors.Is(err, credentials.ErrWrongPassphrase) {
			return nil, nil, err
		}
		return nil, nil, fmt.Errorf("get token for instance %q: %w", name, err)
	}

	client, err := rancher.NewClient(inst.URL, token, inst.InsecureSkipVerify)
	if err != nil {
		return nil, nil, err
	}

	return inst, client, nil
}
