package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/ahoz/kubectl-sheep/internal/instance"
	"github.com/ahoz/kubectl-sheep/internal/kubeconfig"
	"github.com/ahoz/kubectl-sheep/internal/rancher"
	"github.com/spf13/cobra"
)

func newFetchAllCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fetch-all <instance>",
		Short: "Fetch kubeconfigs for all clusters",
		Long:  "Download and store kubeconfigs for every cluster on the given Rancher instance.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			instanceName := args[0]
			merge, _ := cmd.Flags().GetBool("merge")

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
		},
	}

	cmd.Flags().Bool("merge", false, "Merge fetched contexts into ~/.kube/config")

	return cmd
}

func newRefreshAllCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "refresh-all <instance>",
		Short: "Refresh all locally stored kubeconfigs",
		Long:  "Re-fetch kubeconfigs for every cluster that already has a local copy.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			instanceName := args[0]
			merge, _ := cmd.Flags().GetBool("merge")

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
		},
	}

	cmd.Flags().Bool("merge", false, "Merge refreshed contexts into ~/.kube/config")

	return cmd
}
