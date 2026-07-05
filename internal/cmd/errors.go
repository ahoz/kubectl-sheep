package cmd

import (
	"errors"
	"fmt"

	"github.com/ahoz/kubectl-sheep/internal/rancher"
)

func handleRancherError(instanceName string, err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, rancher.ErrTokenInvalid) {
		return fmt.Errorf("rancher token for instance %q is invalid or expired; run: kubectl sheep instance update-token %s", instanceName, instanceName)
	}
	return err
}
