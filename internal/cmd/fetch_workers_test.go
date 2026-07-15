package cmd

import (
	"context"
	"fmt"
	"sync"
	"testing"

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
	results := fetchClusters(context.Background(), "prod", fetcher, clusters)
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
