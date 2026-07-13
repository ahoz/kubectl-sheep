package cmd

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/ahoz/kubectl-sheep/internal/kubeconfig"
	"github.com/ahoz/kubectl-sheep/internal/rancher"
)

type fakeFetcher struct {
	mu   sync.Mutex
	fail map[string]error
}

func (f *fakeFetcher) GenerateKubeconfig(_ context.Context, id string) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if err, ok := f.fail[id]; ok {
		return "", err
	}
	return sampleKubeconfigContent(), nil
}

func TestFetchClustersCollectsErrors(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	clusters := []rancher.Cluster{
		{ID: "c-1", Name: "one"},
		{ID: "c-2", Name: "two"},
		{ID: "c-3", Name: "three"},
	}

	fetcher := &fakeFetcher{fail: map[string]error{"c-2": fmt.Errorf("boom")}}
	results := fetchClustersForTest(context.Background(), "prod", fetcher, clusters)
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	failures := 0
	for _, r := range results {
		if r.err != nil {
			failures++
		}
	}
	if failures != 1 {
		t.Fatalf("expected 1 failure, got %d", failures)
	}
}

func fetchClustersForTest(ctx context.Context, instanceName string, client interface {
	GenerateKubeconfig(context.Context, string) (string, error)
}, clusters []rancher.Cluster) []fetchResult {
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
