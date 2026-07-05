package rancher

import "fmt"

// FindCluster resolves a cluster by ID or name.
func FindCluster(clusters []Cluster, ref string) (*Cluster, error) {
	for i := range clusters {
		if clusters[i].ID == ref || clusters[i].Name == ref {
			return &clusters[i], nil
		}
	}
	return nil, fmt.Errorf("cluster %q not found", ref)
}
