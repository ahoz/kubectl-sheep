package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/ahoz/kubectl-sheep/internal/kubeconfig"
	"github.com/ahoz/kubectl-sheep/internal/rancher"
)

type refreshResult struct {
	cluster rancher.Cluster
	changed bool
	hint    string
	err     error
}

func refreshClusters(ctx context.Context, instanceName string, client kubeconfigGenerator, clusters []rancher.Cluster) []refreshResult {
	return runClusterJobs(clusters, fetchWorkers, func(c rancher.Cluster) refreshResult {
		path, err := kubeconfig.KubeconfigPath(instanceName, c.ID)
		if err != nil {
			return refreshResult{cluster: c, err: err}
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return refreshResult{cluster: c, err: fmt.Errorf("read existing kubeconfig: %w", err)}
		}
		previous := string(data)

		content, err := client.GenerateKubeconfig(ctx, c.ID)
		if err != nil {
			return refreshResult{cluster: c, err: err}
		}
		if err := kubeconfig.Save(instanceName, c.ID, c.Name, content); err != nil {
			return refreshResult{cluster: c, err: err}
		}
		if err := mergeKubeconfig(instanceName, c.Name, content); err != nil {
			return refreshResult{cluster: c, err: err}
		}

		return refreshResult{
			cluster: c,
			changed: previous != content,
			hint:    kubeconfig.TokenExpiryHint(content),
		}
	})
}

func printRefreshResults(w interface {
	Write([]byte) (int, error)
}, instanceName string, results []refreshResult, interactive bool) int {
	configPath, _ := kubeconfigPath()
	failures := 0
	for _, r := range results {
		if r.err != nil {
			failures++
			fprint(w, "ERROR %s (%s): %v\n", r.cluster.Name, r.cluster.ID, r.err)
			continue
		}
		status := "unchanged"
		if r.changed {
			status = "updated"
		}
		contextName := mergeContextName(instanceName, r.cluster.Name)
		if r.hint != "" {
			fprint(w, "OK %s (%s): %s, merged context %q → %s, %s\n", r.cluster.Name, r.cluster.ID, status, contextName, configPath, r.hint)
		} else {
			fprint(w, "OK %s (%s): %s, merged context %q → %s\n", r.cluster.Name, r.cluster.ID, status, contextName, configPath)
		}
		_ = interactive
	}
	return failures
}
