package cmd

import (
	"github.com/spf13/cobra"
)

// Version is injected at build time via -ldflags.
var Version = "dev"

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "version",
		Short:   "Print the version",
		Example: exVersion,
		Run: func(cmd *cobra.Command, args []string) {
			fprintln(cmd.OutOrStdout(), Version)
		},
	}
}
