package kubeconfig

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSaveAndList(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	if err := Save("prod", "c-123", "my-cluster", "apiVersion: v1\n"); err != nil {
		t.Fatalf("Save: %v", err)
	}

	cfgPath, err := KubeconfigPath("prod", "c-123")
	if err != nil {
		t.Fatalf("KubeconfigPath: %v", err)
	}
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatalf("ReadFile kubeconfig: %v", err)
	}
	if string(data) != "apiVersion: v1\n" {
		t.Fatalf("unexpected kubeconfig content: %q", data)
	}

	meta, err := LoadMetadata("prod", "c-123")
	if err != nil {
		t.Fatalf("LoadMetadata: %v", err)
	}
	if meta.ClusterID != "c-123" || meta.ClusterName != "my-cluster" {
		t.Fatalf("unexpected metadata: %+v", meta)
	}
	if time.Since(meta.FetchedAt) > time.Minute {
		t.Fatalf("fetchedAt too old: %v", meta.FetchedAt)
	}

	ids, err := ListStoredClusterIDs("prod")
	if err != nil {
		t.Fatalf("ListStoredClusterIDs: %v", err)
	}
	if len(ids) != 1 || ids[0] != "c-123" {
		t.Fatalf("unexpected ids: %v", ids)
	}

	exists, err := Exists("prod", "c-123")
	if err != nil || !exists {
		t.Fatalf("Exists: %v, %v", exists, err)
	}

	dir := filepath.Join(home, ".kube", sheepDirName, "prod")
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("Stat dir: %v", err)
	}
	if info.Mode().Perm() != 0o700 {
		t.Fatalf("expected dir 0700, got %o", info.Mode().Perm())
	}
}
