package cmd

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/ahoz/kubectl-sheep/internal/instance"
	"github.com/ahoz/kubectl-sheep/internal/kubeconfig"
	"github.com/ahoz/kubectl-sheep/internal/prompt"
	"github.com/ahoz/kubectl-sheep/internal/rancher"
	"github.com/spf13/cobra"
)

func newClusterCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cluster",
		Short: "Manage Rancher clusters",
		Long:  "List, fetch, and refresh kubeconfigs for clusters on a Rancher instance.",
	}

	cmd.AddCommand(newClusterListCmd())
	cmd.AddCommand(newClusterGetCmd())
	cmd.AddCommand(newClusterRefreshCmd())

	return cmd
}

func newClusterListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list <instance>",
		Short: "List clusters for a Rancher instance",
		Long:  "Display all clusters registered on the given Rancher instance.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			_, client, err := instance.RancherClient(args[0])
			if err != nil {
				return err
			}

			clusters, err := client.ListClusters(context.Background())
			if err != nil {
				return handleRancherError(args[0], err)
			}

			if len(clusters) == 0 {
				fprintln(cmd.OutOrStdout(), "No clusters found.")
				return nil
			}

			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fprintln(w, "ID\tNAME\tSTATE")
			for _, c := range clusters {
				fprint(w, "%s\t%s\t%s\n", c.ID, c.Name, c.State)
			}
			return w.Flush()
		},
	}
}

func newClusterGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <instance> <cluster>",
		Short: "Fetch kubeconfig for a single cluster",
		Long:  "Download and store the kubeconfig for the specified cluster.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			instanceName := args[0]
			clusterRef := args[1]
			merge, _ := cmd.Flags().GetBool("merge")
			replace, _ := cmd.Flags().GetBool("replace")
			prefix, _ := cmd.Flags().GetString("prefix")
			contextName, _ := cmd.Flags().GetString("context-name")

			return fetchSingleCluster(cmd, instanceName, clusterRef, mergeOpts{
				merge:       merge,
				replace:     replace,
				prefix:      prefix,
				contextName: contextName,
			}, false)
		},
	}

	cmd.Flags().Bool("merge", false, "Merge into ~/.kube/config without prompting")
	cmd.Flags().Bool("replace", false, "Replace an existing context in ~/.kube/config without prompting (use with --merge)")
	cmd.Flags().String("prefix", "", "Context name prefix to use when merging (default: instance name)")
	cmd.Flags().String("context-name", "", "Exact context name to use when merging")

	return cmd
}

func newClusterRefreshCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "refresh <instance> <cluster>",
		Short: "Refresh kubeconfig for a single cluster",
		Long:  "Re-fetch and update the locally stored kubeconfig for the specified cluster.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			instanceName := args[0]
			clusterRef := args[1]

			exists, err := localClusterExists(instanceName, clusterRef)
			if err != nil {
				return err
			}
			if !exists {
				return fmt.Errorf("no local kubeconfig found for cluster %q; use cluster get first", clusterRef)
			}

			return fetchSingleCluster(cmd, instanceName, clusterRef, mergeOpts{}, true)
		},
	}
}

func localClusterExists(instanceName, clusterRef string) (bool, error) {
	ids, err := kubeconfig.ListStoredClusterIDs(instanceName)
	if err != nil {
		return false, err
	}
	for _, id := range ids {
		if id == clusterRef {
			return true, nil
		}
	}

	_, client, err := instance.RancherClient(instanceName)
	if err != nil {
		return false, err
	}
	clusters, err := client.ListClusters(context.Background())
	if err != nil {
		return false, handleRancherError(instanceName, err)
	}
	cluster, err := rancher.FindCluster(clusters, clusterRef)
	if err != nil {
		return false, err
	}
	return kubeconfig.Exists(instanceName, cluster.ID)
}

type mergeOpts struct {
	merge       bool
	replace     bool
	prefix      string
	contextName string
}

func fetchSingleCluster(cmd *cobra.Command, instanceName, clusterRef string, merge mergeOpts, refresh bool) error {
	_, client, err := instance.RancherClient(instanceName)
	if err != nil {
		return err
	}

	clusters, err := client.ListClusters(context.Background())
	if err != nil {
		return handleRancherError(instanceName, err)
	}

	cluster, err := rancher.FindCluster(clusters, clusterRef)
	if err != nil {
		return err
	}

	var previous string
	if refresh {
		path, err := kubeconfig.KubeconfigPath(instanceName, cluster.ID)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read existing kubeconfig: %w", err)
		}
		previous = string(data)
	}

	content, err := client.GenerateKubeconfig(context.Background(), cluster.ID)
	if err != nil {
		return handleRancherError(instanceName, err)
	}

	if err := kubeconfig.Save(instanceName, cluster.ID, cluster.Name, content); err != nil {
		return err
	}

	path, err := kubeconfig.KubeconfigPath(instanceName, cluster.ID)
	if err != nil {
		return err
	}

	if refresh {
		changed := previous != content
		hint := kubeconfig.TokenExpiryHint(content)
		if changed {
			msg := fmt.Sprintf("kubeconfig for %q updated", cluster.Name)
			if hint != "" {
				msg += ", " + hint
			}
			fprintln(cmd.OutOrStdout(), msg)
		} else {
			fprint(cmd.OutOrStdout(), "kubeconfig for %q unchanged", cluster.Name)
			if hint != "" {
				fprint(cmd.OutOrStdout(), " (%s)", hint)
			}
			fprintln(cmd.OutOrStdout())
		}
		return nil
	}

	fprint(cmd.OutOrStdout(), "Saved kubeconfig for %q to %s\n", cluster.Name, path)

	return offerMergeKubeconfig(mergePromptOptions{
		Merge:       merge.merge,
		Replace:     merge.replace,
		Prefix:      merge.prefix,
		ContextName: merge.contextName,
		In:          cmd.InOrStdin(),
		Out:         cmd.OutOrStdout(),
		IsTTY:       prompt.IsTerminal(os.Stdin),
	}, instanceName, cluster.Name, content)
}
