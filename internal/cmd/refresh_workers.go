package cmd

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/ahoz/kubectl-sheep/internal/kubeconfig"
	"github.com/ahoz/kubectl-sheep/internal/rancher"
)

type refreshResult struct {
	cluster rancher.Cluster
	changed bool
	hint    string
	err     error
}

func refreshClusters(ctx context.Context, instanceName string, client *rancher.Client, clusters []rancher.Cluster) []refreshResult {
	jobs := make(chan rancher.Cluster)
	results := make(chan refreshResult, len(clusters))

	var wg sync.WaitGroup
	for range fetchWorkers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for c := range jobs {
				path, err := kubeconfig.KubeconfigPath(instanceName, c.ID)
				if err != nil {
					results <- refreshResult{cluster: c, err: err}
					continue
				}

				var previous string
				data, err := os.ReadFile(path)
				if err != nil {
					results <- refreshResult{cluster: c, err: fmt.Errorf("read existing kubeconfig: %w", err)}
					continue
				}
				previous = string(data)

				content, err := client.GenerateKubeconfig(ctx, c.ID)
				if err != nil {
					results <- refreshResult{cluster: c, err: err}
					continue
				}
				if err := kubeconfig.Save(instanceName, c.ID, c.Name, content); err != nil {
					results <- refreshResult{cluster: c, err: err}
					continue
				}
				if err := mergeKubeconfig(instanceName, c.Name, content); err != nil {
					results <- refreshResult{cluster: c, err: err}
					continue
				}

				results <- refreshResult{
					cluster: c,
					changed: previous != content,
					hint:    kubeconfig.TokenExpiryHint(content),
				}
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

	collected := make([]refreshResult, 0, len(clusters))
	for r := range results {
		collected = append(collected, r)
	}
	return collected
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
