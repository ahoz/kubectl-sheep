package cmd

import (
	"sync"

	"github.com/ahoz/kubectl-sheep/internal/rancher"
)

const fetchWorkers = 5

func runClusterJobs[T any](clusters []rancher.Cluster, workers int, job func(rancher.Cluster) T) []T {
	if len(clusters) == 0 {
		return nil
	}
	if workers <= 0 {
		workers = 1
	}

	jobs := make(chan rancher.Cluster)
	results := make(chan T, len(clusters))

	var wg sync.WaitGroup
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for c := range jobs {
				results <- job(c)
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

	collected := make([]T, 0, len(clusters))
	for r := range results {
		collected = append(collected, r)
	}
	return collected
}
