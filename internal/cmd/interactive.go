package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/ahoz/kubectl-sheep/internal/config"
	"github.com/ahoz/kubectl-sheep/internal/instance"
	"github.com/ahoz/kubectl-sheep/internal/kubeconfig"
	"github.com/ahoz/kubectl-sheep/internal/prompt"
	"github.com/spf13/cobra"
)

func isInteractive(cmd *cobra.Command) bool {
	noInput, _ := cmd.Root().PersistentFlags().GetBool("no-input")
	return prompt.IsTerminal(os.Stdin) && !noInput
}

func promptRancherInstanceName(cmd *cobra.Command, args []string) (string, error) {
	if len(args) >= 1 {
		return args[0], nil
	}
	if !isInteractive(cmd) {
		return "", fmt.Errorf("rancher-instance name is required")
	}
	name, err := prompt.ReadString(cmd.InOrStdin(), cmd.OutOrStdout(), "Rancher instance name", "")
	if err != nil {
		return "", err
	}
	if name == "" {
		return "", fmt.Errorf("rancher-instance name must not be empty")
	}
	return name, nil
}

func promptRancherInstance(cmd *cobra.Command, args []string) (string, error) {
	if len(args) >= 1 {
		return args[0], nil
	}
	if !isInteractive(cmd) {
		return "", fmt.Errorf("rancher-instance is required")
	}

	cfg, err := config.Load()
	if err != nil {
		return "", err
	}
	if len(cfg.Instances) == 0 {
		return "", fmt.Errorf("no rancher-instances configured; run: kubectl sheep rancher-instance add")
	}

	options := make([]prompt.Choice, len(cfg.Instances))
	for i, inst := range cfg.Instances {
		options[i] = prompt.Choice{Title: inst.Name, Subtitle: inst.URL}
	}

	idx, free, err := prompt.Choose(cmd.InOrStdin(), cmd.OutOrStdout(), "Rancher instance", options)
	if err != nil {
		return "", err
	}
	if idx >= 0 {
		return cfg.Instances[idx].Name, nil
	}
	return free, nil
}

func promptCluster(cmd *cobra.Command, instanceName string) (string, error) {
	_, client, err := instance.RancherClient(instanceName)
	if err != nil {
		return "", err
	}

	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	clusters, err := client.ListClusters(ctx)
	if err != nil {
		return "", handleRancherError(instanceName, err)
	}
	if len(clusters) == 0 {
		return "", fmt.Errorf("no clusters found on rancher-instance %q", instanceName)
	}

	options := make([]prompt.Choice, len(clusters))
	for i, c := range clusters {
		options[i] = prompt.Choice{
			Title:    c.Name,
			Subtitle: fmt.Sprintf("%s · %s", c.ID, c.State),
		}
	}

	idx, free, err := prompt.Choose(cmd.InOrStdin(), cmd.OutOrStdout(), "Cluster", options)
	if err != nil {
		return "", err
	}
	if idx >= 0 {
		return clusters[idx].Name, nil
	}
	return free, nil
}

func promptStoredCluster(cmd *cobra.Command, instanceName string) (string, error) {
	items, err := listStoredClustersWithMeta(instanceName)
	if err != nil {
		return "", err
	}
	if len(items) == 0 {
		return "", fmt.Errorf("no locally stored kubeconfigs for rancher-instance %q", instanceName)
	}

	options := make([]prompt.Choice, len(items))
	for i, item := range items {
		options[i] = prompt.Choice{
			Title:    item.name,
			Subtitle: fmt.Sprintf("%s · fetched %s", item.id, item.fetchedAt),
		}
	}

	idx, free, err := prompt.Choose(cmd.InOrStdin(), cmd.OutOrStdout(), "Stored kubeconfig", options)
	if err != nil {
		return "", err
	}
	if idx >= 0 {
		return items[idx].id, nil
	}
	return free, nil
}

func promptFetchScope(cmd *cobra.Command) (all bool, err error) {
	idx, _, err := prompt.Select(cmd.InOrStdin(), cmd.OutOrStdout(), "Fetch scope", []string{
		"One cluster",
		"All clusters",
	})
	if err != nil {
		return false, err
	}
	return idx == 1, nil
}

func promptRefreshScope(cmd *cobra.Command) (all bool, err error) {
	idx, _, err := prompt.Select(cmd.InOrStdin(), cmd.OutOrStdout(), "Refresh scope", []string{
		"One stored kubeconfig",
		"All stored kubeconfigs",
	})
	if err != nil {
		return false, err
	}
	return idx == 1, nil
}

type storedClusterOption struct {
	id        string
	name      string
	fetchedAt string
}

func listStoredClustersWithMeta(instanceName string) ([]storedClusterOption, error) {
	ids, err := kubeconfig.ListStoredClusterIDs(instanceName)
	if err != nil {
		return nil, err
	}

	var items []storedClusterOption
	for _, id := range ids {
		item := storedClusterOption{id: id, name: "-", fetchedAt: "-"}
		meta, err := kubeconfig.LoadMetadata(instanceName, id)
		if err == nil {
			item.name = meta.ClusterName
			item.fetchedAt = meta.FetchedAt.Format("2006-01-02 15:04")
		}
		items = append(items, item)
	}
	return items, nil
}

func resolveAddNameAndURL(cmd *cobra.Command, args []string) (name, rancherURL string, err error) {
	switch len(args) {
	case 2:
		return args[0], args[1], nil
	case 1:
		if !isInteractive(cmd) {
			return "", "", fmt.Errorf("rancher URL is required: kubectl sheep rancher-instance add %s <url>", args[0])
		}
		url, err := prompt.ReadString(cmd.InOrStdin(), cmd.OutOrStdout(), "Rancher URL", "")
		if err != nil {
			return "", "", err
		}
		if strings.TrimSpace(url) == "" {
			return "", "", fmt.Errorf("rancher URL must not be empty")
		}
		return args[0], url, nil
	case 0:
		if !isInteractive(cmd) {
			return "", "", fmt.Errorf("usage: kubectl sheep rancher-instance add <name> <url>")
		}
		name, err = prompt.ReadString(cmd.InOrStdin(), cmd.OutOrStdout(), "Rancher instance name", "")
		if err != nil {
			return "", "", err
		}
		if name == "" {
			return "", "", fmt.Errorf("rancher-instance name must not be empty")
		}
		url, err := prompt.ReadString(cmd.InOrStdin(), cmd.OutOrStdout(), "Rancher URL", "")
		if err != nil {
			return "", "", err
		}
		if strings.TrimSpace(url) == "" {
			return "", "", fmt.Errorf("rancher URL must not be empty")
		}
		return name, url, nil
	default:
		return "", "", fmt.Errorf("too many arguments")
	}
}

func promptAddOptions(cmd *cobra.Command, fullWizard bool, storage string, insecure bool) (string, bool, error) {
	if !fullWizard || !isInteractive(cmd) {
		return storage, insecure, nil
	}

	storageChoice, err := prompt.ReadString(cmd.InOrStdin(), cmd.OutOrStdout(), "Token storage (encrypted/plaintext)", storage)
	if err != nil {
		return "", false, err
	}
	if err := config.ValidateStorage(storageChoice); err != nil {
		return "", false, err
	}

	skipTLS, err := prompt.Confirm(cmd.InOrStdin(), cmd.OutOrStdout(), "Skip TLS certificate verification", insecure)
	if err != nil {
		return "", false, err
	}

	return storageChoice, skipTLS, nil
}

func promptOpenBrowser(cmd *cobra.Command, openFlag bool) (bool, error) {
	if openFlag || !isInteractive(cmd) {
		return openFlag, nil
	}
	return prompt.Confirm(cmd.InOrStdin(), cmd.OutOrStdout(), "Open Rancher API key page in browser", false)
}

func resolveKubeconfigTarget(cmd *cobra.Command, args []string, needCluster bool) (instanceName, clusterRef string, err error) {
	instanceName, err = promptRancherInstance(cmd, args)
	if err != nil {
		return "", "", err
	}

	if !needCluster {
		return instanceName, "", nil
	}

	if len(args) >= 2 {
		return instanceName, args[1], nil
	}
	if !isInteractive(cmd) {
		return "", "", fmt.Errorf("cluster is required")
	}

	clusterRef, err = promptCluster(cmd, instanceName)
	if err != nil {
		return "", "", err
	}
	return instanceName, clusterRef, nil
}
