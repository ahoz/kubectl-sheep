package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMergeKubeconfigReportsContextName(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	content := sampleKubeconfigContent()
	if err := mergeKubeconfig("prod", "my-cluster", content); err != nil {
		t.Fatalf("mergeKubeconfig: %v", err)
	}

	path := filepath.Join(home, ".kube", "config")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if !strings.Contains(string(data), "prod-my-cluster") {
		t.Fatalf("expected context prod-my-cluster in merged config")
	}
}

func TestReportKubeconfigSavedInteractive(t *testing.T) {
	var out bytes.Buffer
	reportKubeconfigSaved(&out, true, "production", "/tmp/kube.yaml", "prod-production", "/home/user/.kube/config")
	text := out.String()
	if !strings.Contains(text, "Saved and merged") || !strings.Contains(text, "prod-production") {
		t.Fatalf("unexpected output: %s", text)
	}
}

func sampleKubeconfigContent() string {
	return `apiVersion: v1
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
users:
- name: cluster-a
  user:
    token: abc
`
}
