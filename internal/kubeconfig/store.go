package kubeconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

const sheepDirName = "sheep"

// Metadata tracks when a kubeconfig was fetched.
type Metadata struct {
	FetchedAt   time.Time `yaml:"fetchedAt"`
	ClusterID   string    `yaml:"clusterId"`
	ClusterName string    `yaml:"clusterName"`
}

// StoreDir returns ~/.kube/sheep/<instance>.
func StoreDir(instance string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}
	return filepath.Join(home, ".kube", sheepDirName, instance), nil
}

// KubeconfigPath returns the path for a cluster kubeconfig file.
func KubeconfigPath(instance, clusterID string) (string, error) {
	dir, err := StoreDir(instance)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, clusterID+".yaml"), nil
}

// MetadataPath returns the path for a cluster metadata file.
func MetadataPath(instance, clusterID string) (string, error) {
	dir, err := StoreDir(instance)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, clusterID+".meta.yaml"), nil
}

// Save writes kubeconfig and metadata for a cluster.
func Save(instance, clusterID, clusterName, content string) error {
	dir, err := StoreDir(instance)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create kubeconfig directory: %w", err)
	}

	cfgPath, err := KubeconfigPath(instance, clusterID)
	if err != nil {
		return err
	}
	if err := os.WriteFile(cfgPath, []byte(content), 0o600); err != nil {
		return fmt.Errorf("write kubeconfig: %w", err)
	}

	meta := Metadata{
		FetchedAt:   time.Now().UTC(),
		ClusterID:   clusterID,
		ClusterName: clusterName,
	}
	metaPath, err := MetadataPath(instance, clusterID)
	if err != nil {
		return err
	}
	data, err := yaml.Marshal(meta)
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}
	if err := os.WriteFile(metaPath, data, 0o600); err != nil {
		return fmt.Errorf("write metadata: %w", err)
	}
	return nil
}

// LoadMetadata reads metadata for a stored cluster.
func LoadMetadata(instance, clusterID string) (*Metadata, error) {
	metaPath, err := MetadataPath(instance, clusterID)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(metaPath)
	if err != nil {
		return nil, fmt.Errorf("read metadata: %w", err)
	}
	var meta Metadata
	if err := yaml.Unmarshal(data, &meta); err != nil {
		return nil, fmt.Errorf("parse metadata: %w", err)
	}
	return &meta, nil
}

// ListStoredClusterIDs returns cluster IDs that have local kubeconfig files.
func ListStoredClusterIDs(instance string) ([]string, error) {
	dir, err := StoreDir(instance)
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read kubeconfig directory: %w", err)
	}

	var ids []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if filepath.Ext(name) == ".yaml" && filepath.Base(name) != "" {
			base := name[:len(name)-len(".yaml")]
			if filepath.Ext(base) == ".meta" {
				continue
			}
			ids = append(ids, base)
		}
	}
	return ids, nil
}

// Exists reports whether a kubeconfig file exists for the cluster.
func Exists(instance, clusterID string) (bool, error) {
	path, err := KubeconfigPath(instance, clusterID)
	if err != nil {
		return false, err
	}
	_, err = os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
