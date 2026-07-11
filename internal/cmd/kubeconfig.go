package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/ahoz/kubectl-sheep/internal/instance"
	"github.com/ahoz/kubectl-sheep/internal/kubeconfig"
	"github.com/ahoz/kubectl-sheep/internal/prompt"
	"github.com/ahoz/kubectl-sheep/internal/rancher"
	"github.com/spf13/cobra"
)

func newKubeconfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "kubeconfig",
		Short:   "Manage locally stored kubeconfigs",
		Long:    "Fetch, refresh, list, and install kubeconfigs from Rancher-managed clusters.",
		Example: exKubeconfig,
	}

	cmd.AddCommand(newKubeconfigListCmd())
	cmd.AddCommand(newKubeconfigGetCmd())
	cmd.AddCommand(newKubeconfigFetchCmd())
	cmd.AddCommand(newKubeconfigRefreshCmd())
	cmd.AddCommand(newKubeconfigInstallExecCmd())

	return cmd
}

func newKubeconfigListCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "list <rancher-instance>",
		Short:   "List locally stored kubeconfigs",
		Long:    "Display kubeconfigs already downloaded for the given Rancher instance.",
		Example: exKubeconfigList,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			instanceName := args[0]

			ids, err := kubeconfig.ListStoredClusterIDs(instanceName)
			if err != nil {
				return err
			}
			if len(ids) == 0 {
				fprintln(cmd.OutOrStdout(), "No locally stored kubeconfigs.")
				return nil
			}

			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fprintln(w, "CLUSTER ID\tCLUSTER NAME\tFETCHED AT\tPATH")
			for _, id := range ids {
				meta, err := kubeconfig.LoadMetadata(instanceName, id)
				if err != nil {
					fprint(w, "%s\t-\t-\t-\n", id)
					continue
				}
				path, err := kubeconfig.KubeconfigPath(instanceName, id)
				if err != nil {
					path = "-"
				}
				fprint(w, "%s\t%s\t%s\t%s\n", id, meta.ClusterName, meta.FetchedAt.Format("2006-01-02 15:04:05"), path)
			}
			return w.Flush()
		},
	}
}

func newKubeconfigGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "get [rancher-instance] [cluster]",
		Short:   "Fetch kubeconfig for a single cluster",
		Long:    "Download and store the kubeconfig for the specified cluster.",
		Example: exKubeconfigGet,
		Args:    cobra.MaximumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if isInteractive(cmd) && len(args) < 2 {
				prompt.Intro(cmd.OutOrStdout(), "Fetch cluster kubeconfig")
			}

			instanceName, clusterRef, err := resolveKubeconfigTarget(cmd, args, true)
			if err != nil {
				return err
			}

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

	addKubeconfigMergeFlags(cmd)

	return cmd
}

func newKubeconfigFetchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "fetch [rancher-instance] [cluster]",
		Short:   "Fetch kubeconfigs from Rancher",
		Long:    "Download and store kubeconfigs for one cluster or for all clusters with --all.",
		Example: exKubeconfigFetch,
		Args:    cobra.MaximumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runKubeconfigFetch(cmd, args)
		},
	}

	cmd.Flags().Bool("all", false, "Fetch kubeconfigs for all clusters on the Rancher instance")
	addKubeconfigMergeFlags(cmd)

	return cmd
}

func runKubeconfigFetch(cmd *cobra.Command, args []string) error {
	all, _ := cmd.Flags().GetBool("all")
	merge, _ := cmd.Flags().GetBool("merge")

	if all && len(args) == 2 {
		return fmt.Errorf("use either a cluster argument or --all, not both")
	}

	if isInteractive(cmd) && len(args) < 2 && !all {
		prompt.Intro(cmd.OutOrStdout(), "Fetch kubeconfigs from Rancher")
	}

	instanceName, err := promptRancherInstance(cmd, args)
	if err != nil {
		return err
	}

	if len(args) == 2 {
		replace, _ := cmd.Flags().GetBool("replace")
		prefix, _ := cmd.Flags().GetString("prefix")
		contextName, _ := cmd.Flags().GetString("context-name")
		return fetchSingleCluster(cmd, instanceName, args[1], mergeOpts{
			merge:       merge,
			replace:     replace,
			prefix:      prefix,
			contextName: contextName,
		}, false)
	}

	if all {
		return runFetchAll(cmd, instanceName, merge)
	}

	if len(args) == 1 && !isInteractive(cmd) {
		return fmt.Errorf("specify a cluster or pass --all")
	}

	if isInteractive(cmd) {
		scopeAll, err := promptFetchScope(cmd)
		if err != nil {
			return err
		}
		if scopeAll {
			return runFetchAll(cmd, instanceName, merge)
		}
		clusterRef, err := promptCluster(cmd, instanceName)
		if err != nil {
			return err
		}
		replace, _ := cmd.Flags().GetBool("replace")
		prefix, _ := cmd.Flags().GetString("prefix")
		contextName, _ := cmd.Flags().GetString("context-name")
		return fetchSingleCluster(cmd, instanceName, clusterRef, mergeOpts{
			merge:       merge,
			replace:     replace,
			prefix:      prefix,
			contextName: contextName,
		}, false)
	}

	return fmt.Errorf("specify a cluster or pass --all")
}

func newKubeconfigRefreshCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "refresh [rancher-instance] [cluster]",
		Short:   "Refresh locally stored kubeconfigs",
		Long:    "Re-fetch kubeconfigs for one cluster or for all locally stored clusters with --all.",
		Example: exKubeconfigRefresh,
		Args:    cobra.MaximumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runKubeconfigRefresh(cmd, args)
		},
	}

	cmd.Flags().Bool("all", false, "Refresh all locally stored kubeconfigs for the Rancher instance")
	cmd.Flags().Bool("merge", false, "Merge refreshed contexts into ~/.kube/config")

	return cmd
}

func runKubeconfigRefresh(cmd *cobra.Command, args []string) error {
	all, _ := cmd.Flags().GetBool("all")
	merge, _ := cmd.Flags().GetBool("merge")

	if all && len(args) == 2 {
		return fmt.Errorf("use either a cluster argument or --all, not both")
	}

	if isInteractive(cmd) && len(args) < 2 && !all {
		prompt.Intro(cmd.OutOrStdout(), "Refresh locally stored kubeconfigs")
	}

	instanceName, err := promptRancherInstance(cmd, args)
	if err != nil {
		return err
	}

	if len(args) == 2 {
		clusterRef := args[1]
		exists, err := localClusterExists(instanceName, clusterRef)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("no local kubeconfig found for cluster %q; use kubeconfig get or fetch first", clusterRef)
		}
		return fetchSingleCluster(cmd, instanceName, clusterRef, mergeOpts{}, true)
	}

	if all {
		return runRefreshAll(cmd, instanceName, merge)
	}

	if len(args) == 1 && !isInteractive(cmd) {
		return fmt.Errorf("specify a cluster or pass --all")
	}

	if isInteractive(cmd) {
		scopeAll, err := promptRefreshScope(cmd)
		if err != nil {
			return err
		}
		if scopeAll {
			return runRefreshAll(cmd, instanceName, merge)
		}
		clusterRef, err := promptStoredCluster(cmd, instanceName)
		if err != nil {
			return err
		}
		return fetchSingleCluster(cmd, instanceName, clusterRef, mergeOpts{}, true)
	}

	return fmt.Errorf("specify a cluster or pass --all")
}

func newKubeconfigInstallExecCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "install-exec <rancher-instance> <cluster>",
		Short:   "Install an exec-based kubeconfig context",
		Long:    "Create or update a kubeconfig context that calls kubectl-sheep to load credentials on demand.",
		Example: exKubeconfigInstallExec,
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			replace, _ := cmd.Flags().GetBool("replace")
			prefix, _ := cmd.Flags().GetString("prefix")
			contextName, _ := cmd.Flags().GetString("context-name")
			execCommand, _ := cmd.Flags().GetString("exec-command")

			return installExecCluster(cmd, args[0], args[1], execInstallOpts{
				replace:     replace,
				prefix:      prefix,
				contextName: contextName,
				execCommand: execCommand,
			})
		},
	}

	cmd.Flags().Bool("replace", false, "Replace an existing context in ~/.kube/config without prompting")
	cmd.Flags().String("prefix", "", "Context name prefix to use (default: rancher-instance name)")
	cmd.Flags().String("context-name", "", "Exact context name to use")
	cmd.Flags().String("exec-command", "kubectl-sheep", "Command used by kubeconfig exec users")

	return cmd
}

func addKubeconfigMergeFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("merge", false, "Merge into ~/.kube/config without prompting")
	cmd.Flags().Bool("replace", false, "Replace an existing context in ~/.kube/config without prompting (use with --merge)")
	cmd.Flags().String("prefix", "", "Context name prefix to use when merging (default: rancher-instance name)")
	cmd.Flags().String("context-name", "", "Exact context name to use when merging")
}

func runFetchAll(cmd *cobra.Command, instanceName string, merge bool) error {
	_, client, err := instance.RancherClient(instanceName)
	if err != nil {
		return err
	}

	clusters, err := client.ListClusters(context.Background())
	if err != nil {
		return handleRancherError(instanceName, err)
	}
	if len(clusters) == 0 {
		fprintln(cmd.OutOrStdout(), "No clusters found.")
		return nil
	}

	results := fetchClusters(context.Background(), instanceName, client, clusters, merge)
	failures := printFetchResults(cmd.OutOrStdout(), instanceName, results)
	if failures > 0 {
		for _, r := range results {
			if errors.Is(r.err, rancher.ErrTokenInvalid) {
				return handleRancherError(instanceName, rancher.ErrTokenInvalid)
			}
		}
		return fmt.Errorf("%d of %d clusters failed", failures, len(clusters))
	}
	return nil
}

func runRefreshAll(cmd *cobra.Command, instanceName string, merge bool) error {
	storedIDs, err := kubeconfig.ListStoredClusterIDs(instanceName)
	if err != nil {
		return err
	}
	if len(storedIDs) == 0 {
		fprintln(cmd.OutOrStdout(), "No locally stored kubeconfigs.")
		return nil
	}

	_, client, err := instance.RancherClient(instanceName)
	if err != nil {
		return err
	}

	allClusters, err := client.ListClusters(context.Background())
	if err != nil {
		return err
	}

	byID := make(map[string]rancher.Cluster, len(allClusters))
	for _, c := range allClusters {
		byID[c.ID] = c
	}

	var toRefresh []rancher.Cluster
	for _, id := range storedIDs {
		c, ok := byID[id]
		if !ok {
			fprint(cmd.OutOrStdout(), "SKIP %s: cluster no longer exists in Rancher\n", id)
			continue
		}
		toRefresh = append(toRefresh, c)
	}

	if len(toRefresh) == 0 {
		return nil
	}

	results := refreshClusters(context.Background(), instanceName, client, toRefresh, merge)
	failures := printRefreshResults(cmd.OutOrStdout(), results)
	if failures > 0 {
		for _, r := range results {
			if errors.Is(r.err, rancher.ErrTokenInvalid) {
				return handleRancherError(instanceName, rancher.ErrTokenInvalid)
			}
		}
		return fmt.Errorf("%d of %d clusters failed", failures, len(toRefresh))
	}
	return nil
}

type execInstallOpts struct {
	replace     bool
	prefix      string
	contextName string
	execCommand string
}

func installExecCluster(cmd *cobra.Command, instanceName, clusterRef string, opts execInstallOpts) error {
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

	content, err := client.GenerateKubeconfig(context.Background(), cluster.ID)
	if err != nil {
		return handleRancherError(instanceName, err)
	}

	if err := kubeconfig.Save(instanceName, cluster.ID, cluster.Name, content); err != nil {
		return err
	}

	contextName := mergeContextName(instanceName, cluster.Name, opts.prefix, opts.contextName)
	execContent, err := buildExecKubeconfig(content, execKubeconfigOptions{
		contextName: contextName,
		command:     opts.execCommand,
		args:        []string{"auth", "exec", instanceName, cluster.ID},
	})
	if err != nil {
		return err
	}

	return offerMergeKubeconfig(mergePromptOptions{
		Merge:       true,
		Replace:     opts.replace,
		ContextName: contextName,
		In:          cmd.InOrStdin(),
		Out:         cmd.OutOrStdout(),
		IsTTY:       isInteractive(cmd),
	}, instanceName, cluster.Name, execContent)
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

	if isInteractive(cmd) {
		prompt.Success(cmd.OutOrStdout(), fmt.Sprintf(`Saved kubeconfig for %q`, cluster.Name))
		prompt.Note(cmd.OutOrStdout(), path)
	} else {
		fprint(cmd.OutOrStdout(), "Saved kubeconfig for %q to %s\n", cluster.Name, path)
	}

	return offerMergeKubeconfig(mergePromptOptions{
		Merge:       merge.merge,
		Replace:     merge.replace,
		Prefix:      merge.prefix,
		ContextName: merge.contextName,
		In:          cmd.InOrStdin(),
		Out:         cmd.OutOrStdout(),
		IsTTY:       isInteractive(cmd),
	}, instanceName, cluster.Name, content)
}
