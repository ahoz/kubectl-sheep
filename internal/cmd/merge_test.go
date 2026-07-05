package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMergeKubeconfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	content := `apiVersion: v1
kind: Config
clusters:
- name: cluster-a
  cluster:
    server: https://example.com
contexts:
- name: cluster-a
  context:
    cluster: cluster-a
    user: cluster-a
current-context: cluster-a
users:
- name: cluster-a
  user:
    token: abc
`

	if err := mergeKubeconfig("prod", "my-cluster", content); err != nil {
		t.Fatalf("mergeKubeconfig: %v", err)
	}

	path := filepath.Join(home, ".kube", "config")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	text := string(data)
	if !strings.Contains(text, "prod-my-cluster") {
		t.Fatalf("expected prefixed context name in merged config, got:\n%s", text)
	}
	if !strings.Contains(text, "https://example.com") {
		t.Fatalf("expected server in merged config")
	}
}

func TestMergeKubeconfigCollisionOverwrite(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	kubeDir := filepath.Join(home, ".kube")
	if err := os.MkdirAll(kubeDir, 0o700); err != nil {
		t.Fatal(err)
	}

	existing := `apiVersion: v1
kind: Config
clusters:
- name: prod-my-cluster
  cluster:
    server: https://old.example.com
contexts:
- name: prod-my-cluster
  context:
    cluster: prod-my-cluster
    user: prod-my-cluster
current-context: prod-my-cluster
users:
- name: prod-my-cluster
  user:
    token: old
`
	if err := os.WriteFile(filepath.Join(kubeDir, "config"), []byte(existing), 0o600); err != nil {
		t.Fatal(err)
	}

	content := `apiVersion: v1
kind: Config
clusters:
- name: cluster-a
  cluster:
    server: https://new.example.com
contexts:
- name: cluster-a
  context:
    cluster: cluster-a
    user: cluster-a
users:
- name: cluster-a
  user:
    token: new
`
	if err := mergeKubeconfig("prod", "my-cluster", content); err != nil {
		t.Fatalf("mergeKubeconfig: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(kubeDir, "config"))
	if !strings.Contains(string(data), "https://new.example.com") {
		t.Fatalf("expected overwrite with new server, got:\n%s", data)
	}
}
