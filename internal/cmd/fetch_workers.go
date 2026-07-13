package cmd

import (
	"context"
	"sync"

	"github.com/ahoz/kubectl-sheep/internal/kubeconfig"
	"github.com/ahoz/kubectl-sheep/internal/rancher"
)

const fetchWorkers = 5

type fetchResult struct {
	cluster rancher.Cluster
	err     error
}

func fetchClusters(ctx context.Context, instanceName string, client *rancher.Client, clusters []rancher.Cluster) []fetchResult {
	jobs := make(chan rancher.Cluster)
	results := make(chan fetchResult, len(clusters))

	var wg sync.WaitGroup
	for range fetchWorkers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for c := range jobs {
				content, err := client.GenerateKubeconfig(ctx, c.ID)
				if err != nil {
					results <- fetchResult{cluster: c, err: err}
					continue
				}
				if err := kubeconfig.Save(instanceName, c.ID, c.Name, content); err != nil {
					results <- fetchResult{cluster: c, err: err}
					continue
				}
				if err := mergeKubeconfig(instanceName, c.Name, content); err != nil {
					results <- fetchResult{cluster: c, err: err}
					continue
				}
				results <- fetchResult{cluster: c, err: nil}
			}
		}()
	}

	go func() {
		for _, c := range clusters {
			jobs <- c
		}
		close(jobs)
		wg.Wait()
		close(results)
	}()

	collected := make([]fetchResult, 0, len(clusters))
	for r := range results {
		collected = append(collected, r)
	}
	return collected
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
