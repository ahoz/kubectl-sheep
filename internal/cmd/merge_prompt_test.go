package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOfferMergeKubeconfigInteractiveYes(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	var out bytes.Buffer
	err := offerMergeKubeconfig(mergePromptOptions{
		In:    strings.NewReader("y\n"),
		Out:   &out,
		IsTTY: true,
	}, "prod", "my-cluster", sampleKubeconfigContent())
	if err != nil {
		t.Fatalf("offerMergeKubeconfig: %v", err)
	}
	if !strings.Contains(out.String(), `Merged context "prod-my-cluster"`) {
		t.Fatalf("unexpected output: %s", out.String())
	}
}

func TestOfferMergeKubeconfigInteractiveNo(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	var out bytes.Buffer
	err := offerMergeKubeconfig(mergePromptOptions{
		In:    strings.NewReader("n\n"),
		Out:   &out,
		IsTTY: true,
	}, "prod", "my-cluster", sampleKubeconfigContent())
	if err != nil {
		t.Fatalf("offerMergeKubeconfig: %v", err)
	}
	if _, err := os.Stat(filepath.Join(home, ".kube", "config")); !os.IsNotExist(err) {
		t.Fatal("expected no kubeconfig merge")
	}
}

func TestOfferMergeKubeconfigReplacePrompt(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	kubeDir := filepath.Join(home, ".kube")
	_ = os.MkdirAll(kubeDir, 0o700)
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
users:
- name: prod-my-cluster
  user:
    token: old
`
	_ = os.WriteFile(filepath.Join(kubeDir, "config"), []byte(existing), 0o600)

	var out bytes.Buffer
	err := offerMergeKubeconfig(mergePromptOptions{
		In:    strings.NewReader("y\ny\n"),
		Out:   &out,
		IsTTY: true,
	}, "prod", "my-cluster", sampleKubeconfigContent())
	if err != nil {
		t.Fatalf("offerMergeKubeconfig: %v", err)
	}
	data, _ := os.ReadFile(filepath.Join(kubeDir, "config"))
	if !strings.Contains(string(data), "https://example.com") {
		t.Fatalf("expected replaced server, got:\n%s", data)
	}
}

func TestOfferMergeKubeconfigReplaceDecline(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	kubeDir := filepath.Join(home, ".kube")
	_ = os.MkdirAll(kubeDir, 0o700)
	existing := `apiVersion: v1
kind: Config
contexts:
- name: prod-my-cluster
  context:
    cluster: prod-my-cluster
    user: prod-my-cluster
clusters:
- name: prod-my-cluster
  cluster:
    server: https://old.example.com
users:
- name: prod-my-cluster
  user:
    token: old
`
	_ = os.WriteFile(filepath.Join(kubeDir, "config"), []byte(existing), 0o600)

	var out bytes.Buffer
	err := offerMergeKubeconfig(mergePromptOptions{
		In:    strings.NewReader("y\nn\n"),
		Out:   &out,
		IsTTY: true,
	}, "prod", "my-cluster", sampleKubeconfigContent())
	if err != nil {
		t.Fatalf("offerMergeKubeconfig: %v", err)
	}
	if !strings.Contains(out.String(), "Skipped merge") {
		t.Fatalf("unexpected output: %s", out.String())
	}
	data, _ := os.ReadFile(filepath.Join(kubeDir, "config"))
	if !strings.Contains(string(data), "https://old.example.com") {
		t.Fatalf("expected original config preserved, got:\n%s", data)
	}
}

func TestOfferMergeKubeconfigMergeReplaceFlags(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	kubeDir := filepath.Join(home, ".kube")
	_ = os.MkdirAll(kubeDir, 0o700)
	existing := `apiVersion: v1
kind: Config
contexts:
- name: prod-my-cluster
  context:
    cluster: prod-my-cluster
    user: prod-my-cluster
clusters:
- name: prod-my-cluster
  cluster:
    server: https://old.example.com
users:
- name: prod-my-cluster
  user:
    token: old
`
	_ = os.WriteFile(filepath.Join(kubeDir, "config"), []byte(existing), 0o600)

	err := offerMergeKubeconfig(mergePromptOptions{
		Merge:   true,
		Replace: true,
		In:      strings.NewReader(""),
		Out:     &bytes.Buffer{},
		IsTTY:   false,
	}, "prod", "my-cluster", sampleKubeconfigContent())
	if err != nil {
		t.Fatalf("offerMergeKubeconfig: %v", err)
	}
	data, _ := os.ReadFile(filepath.Join(kubeDir, "config"))
	if !strings.Contains(string(data), "https://example.com") {
		t.Fatalf("expected replaced server, got:\n%s", data)
	}
}

func TestOfferMergeKubeconfigMergeCollisionNoTTY(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	kubeDir := filepath.Join(home, ".kube")
	_ = os.MkdirAll(kubeDir, 0o700)
	_ = os.WriteFile(filepath.Join(kubeDir, "config"), []byte(`apiVersion: v1
kind: Config
contexts:
- name: prod-my-cluster
  context:
    cluster: prod-my-cluster
    user: prod-my-cluster
clusters:
- name: prod-my-cluster
  cluster:
    server: https://old.example.com
users:
- name: prod-my-cluster
  user:
    token: old
`), 0o600)

	err := offerMergeKubeconfig(mergePromptOptions{
		Merge: true,
		In:    strings.NewReader(""),
		Out:   &bytes.Buffer{},
		IsTTY: false,
	}, "prod", "my-cluster", sampleKubeconfigContent())
	if err == nil || !strings.Contains(err.Error(), "--replace") {
		t.Fatalf("expected replace hint error, got %v", err)
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
