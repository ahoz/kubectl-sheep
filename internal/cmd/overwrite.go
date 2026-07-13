package cmd

import (
	"fmt"

	"github.com/ahoz/kubectl-sheep/internal/kubeconfig"
	"github.com/ahoz/kubectl-sheep/internal/prompt"
	"github.com/ahoz/kubectl-sheep/internal/rancher"
	"github.com/spf13/cobra"
)

type overwritePolicy int

const (
	overwriteAll overwritePolicy = iota
	overwriteAskEach
)

func promptOverwritePolicy(cmd *cobra.Command) (overwritePolicy, error) {
	idx, _, err := prompt.Select(cmd.InOrStdin(), cmd.OutOrStdout(), "Existing kubeconfigs", []string{
		"Overwrite all existing kubeconfigs",
		"Ask before overwriting each existing kubeconfig",
	})
	if err != nil {
		return overwriteAll, err
	}
	if idx == 1 {
		return overwriteAskEach, nil
	}
	return overwriteAll, nil
}

func resolveOverwritePolicy(cmd *cobra.Command, instanceName string, clusters []rancher.Cluster) (overwritePolicy, error) {
	if !isInteractive(cmd) {
		return overwriteAll, nil
	}

	anyExist, err := anyStoredKubeconfig(instanceName, clusters)
	if err != nil {
		return overwriteAll, err
	}
	if !anyExist {
		return overwriteAll, nil
	}

	return promptOverwritePolicy(cmd)
}

func anyStoredKubeconfig(instanceName string, clusters []rancher.Cluster) (bool, error) {
	for _, c := range clusters {
		exists, err := kubeconfig.Exists(instanceName, c.ID)
		if err != nil {
			return false, err
		}
		if exists {
			return true, nil
		}
	}
	return false, nil
}

func filterClustersForFetch(cmd *cobra.Command, instanceName string, clusters []rancher.Cluster, policy overwritePolicy) ([]rancher.Cluster, error) {
	if policy != overwriteAskEach {
		return clusters, nil
	}

	var filtered []rancher.Cluster
	for _, c := range clusters {
		exists, err := kubeconfig.Exists(instanceName, c.ID)
		if err != nil {
			return nil, err
		}
		if !exists {
			filtered = append(filtered, c)
			continue
		}

		ok, err := prompt.Confirm(cmd.InOrStdin(), cmd.OutOrStdout(),
			fmt.Sprintf(`Overwrite existing kubeconfig for %q`, c.Name), true)
		if err != nil {
			return nil, err
		}
		if ok {
			filtered = append(filtered, c)
			continue
		}
		fprint(cmd.OutOrStdout(), "SKIP %s (%s): keeping existing kubeconfig\n", c.Name, c.ID)
	}
	return filtered, nil
}

func shouldOverwriteSingle(cmd *cobra.Command, instanceName string, cluster rancher.Cluster, policy overwritePolicy) (bool, error) {
	exists, err := kubeconfig.Exists(instanceName, cluster.ID)
	if err != nil {
		return false, err
	}
	if !exists {
		return true, nil
	}
	if policy == overwriteAll {
		return true, nil
	}

	ok, err := prompt.Confirm(cmd.InOrStdin(), cmd.OutOrStdout(),
		fmt.Sprintf(`Overwrite existing kubeconfig for %q`, cluster.Name), true)
	if err != nil {
		return false, err
	}
	return ok, nil
}
