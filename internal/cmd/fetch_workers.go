package cmd

import (
	"context"

	"github.com/ahoz/kubectl-sheep/internal/kubeconfig"
	"github.com/ahoz/kubectl-sheep/internal/rancher"
)

type kubeconfigGenerator interface {
	GenerateKubeconfig(context.Context, string) (string, error)
}

type fetchResult struct {
	cluster rancher.Cluster
	err     error
}

func fetchClusters(ctx context.Context, instanceName string, client kubeconfigGenerator, clusters []rancher.Cluster) []fetchResult {
	return runClusterJobs(clusters, fetchWorkers, func(c rancher.Cluster) fetchResult {
		content, err := client.GenerateKubeconfig(ctx, c.ID)
		if err != nil {
			return fetchResult{cluster: c, err: err}
		}
		if err := kubeconfig.Save(instanceName, c.ID, c.Name, content); err != nil {
			return fetchResult{cluster: c, err: err}
		}
		if err := mergeKubeconfig(instanceName, c.Name, content); err != nil {
			return fetchResult{cluster: c, err: err}
		}
		return fetchResult{cluster: c}
	})
}

func printFetchResults(w interface {
	Write([]byte) (int, error)
}, instanceName string, results []fetchResult, interactive bool) int {
	configPath, _ := kubeconfigPath()
	failures := 0
	for _, r := range results {
		if r.err != nil {
			failures++
			fprint(w, "ERROR %s (%s): %v\n", r.cluster.Name, r.cluster.ID, r.err)
			continue
		}
		path, err := kubeconfig.KubeconfigPath(instanceName, r.cluster.ID)
		if err != nil {
			fprint(w, "OK %s (%s), context %q\n", r.cluster.Name, r.cluster.ID, mergeContextName(instanceName, r.cluster.Name))
			continue
		}
		contextName := mergeContextName(instanceName, r.cluster.Name)
		if interactive {
			fprint(w, "✓ %s: saved to %s, merged context %q → %s\n", r.cluster.Name, path, contextName, configPath)
		} else {
			fprint(w, "OK %s (%s) -> %s, context %q → %s\n", r.cluster.Name, r.cluster.ID, path, contextName, configPath)
		}
	}
	return failures
}
