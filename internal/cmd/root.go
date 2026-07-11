package cmd

import (
	"github.com/spf13/cobra"
)

const (
	shortDescription = "Fetch and manage kubeconfigs from Rancher-managed clusters"
	longDescription  = `A kubectl plugin to manage multiple Rancher instances, list their
downstream clusters, and fetch/refresh kubeconfigs individually or in bulk.
Rancher API tokens can be stored either as plaintext or encrypted
(passphrase-protected file backend), selectable per instance.`
)

// NewRootCmd returns the root command for kubectl sheep.
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "sheep",
		Short: shortDescription,
		Long:  longDescription,
	}

	root.AddCommand(newVersionCmd())
	root.AddCommand(newAuthCmd())
	root.AddCommand(newRancherInstanceCmd())
	root.AddCommand(newKubeconfigCmd())

	return root
}
