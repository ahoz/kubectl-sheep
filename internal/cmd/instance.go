package cmd

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/ahoz/kubectl-sheep/internal/config"
	"github.com/ahoz/kubectl-sheep/internal/credentials"
	"github.com/ahoz/kubectl-sheep/internal/prompt"
	"github.com/ahoz/kubectl-sheep/internal/rancher"
	"github.com/spf13/cobra"
)

func newInstanceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "instance",
		Short: "Manage Rancher instances",
		Long:  "Add, list, remove, and configure Rancher instances.",
	}

	cmd.AddCommand(newInstanceAddCmd())
	cmd.AddCommand(newInstanceListCmd())
	cmd.AddCommand(newInstanceRemoveCmd())
	cmd.AddCommand(newInstanceSetStorageCmd())
	cmd.AddCommand(newInstanceUpdateTokenCmd())

	return cmd
}

func newInstanceAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <name>",
		Short: "Add a Rancher instance",
		Long:  "Register a new Rancher instance with name, URL, and token storage preference.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			url, _ := cmd.Flags().GetString("url")
			storage, _ := cmd.Flags().GetString("storage")
			insecure, _ := cmd.Flags().GetBool("insecure")

			cfg, err := config.Load()
			if err != nil {
				return err
			}

			inst := config.Instance{
				Name:               name,
				URL:                url,
				InsecureSkipVerify: insecure,
				Storage:            storage,
			}
			if err := cfg.AddInstance(inst); err != nil {
				return fmt.Errorf("add instance: %w", err)
			}

			if tokenPageURL, err := rancher.TokenCreatePageURL(url); err != nil {
				return fmt.Errorf("build token page URL: %w", err)
			} else {
				prompt.PrintTokenCreateHint(cmd.OutOrStdout(), tokenPageURL)
			}

			token, err := prompt.ReadSecret(os.Stdin, cmd.OutOrStdout(), "Rancher API token")
			if err != nil {
				return err
			}
			if token == "" {
				return fmt.Errorf("token must not be empty")
			}

			store, err := credentials.NewStore(storage)
			if err != nil {
				return err
			}
			if err := store.Set(name, token); err != nil {
				return err
			}

			if err := cfg.Save(); err != nil {
				_ = store.Delete(name)
				return err
			}

			fprint(cmd.OutOrStdout(), "Added instance %q\n", name)
			return nil
		},
	}

	cmd.Flags().String("url", "", "Rancher server URL (required)")
	cmd.Flags().String("storage", config.StorageEncrypted, "Token storage mode: plaintext or encrypted")
	cmd.Flags().Bool("insecure", false, "Skip TLS certificate verification")

	_ = cmd.MarkFlagRequired("url")

	return cmd
}

func newInstanceListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List configured Rancher instances",
		Long:  "Display all registered Rancher instances and their settings.",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			if len(cfg.Instances) == 0 {
				fprintln(cmd.OutOrStdout(), "No instances configured.")
				return nil
			}

			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fprintln(w, "NAME\tURL\tSTORAGE\tINSECURE")
			for _, inst := range cfg.Instances {
				insecure := "false"
				if inst.InsecureSkipVerify {
					insecure = "true"
				}
				fprint(w, "%s\t%s\t%s\t%s\n", inst.Name, inst.URL, inst.Storage, insecure)
			}
			return w.Flush()
		},
	}
}

func newInstanceRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove a Rancher instance",
		Long:  "Remove a Rancher instance and its stored credentials.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			cfg, err := config.Load()
			if err != nil {
				return err
			}

			inst, err := cfg.Find(name)
			if err != nil {
				return err
			}

			store, err := credentials.NewStore(inst.Storage)
			if err != nil {
				return err
			}
			_ = store.Delete(name)

			if err := cfg.RemoveInstance(name); err != nil {
				return fmt.Errorf("remove instance: %w", err)
			}
			if err := cfg.Save(); err != nil {
				return err
			}

			fprint(cmd.OutOrStdout(), "Removed instance %q\n", name)
			return nil
		},
	}
}

func newInstanceSetStorageCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-storage <name>",
		Short: "Change token storage mode for an instance",
		Long:  "Migrate an instance's Rancher token between plaintext and encrypted storage.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			to, _ := cmd.Flags().GetString("to")

			if err := config.ValidateStorage(to); err != nil {
				return err
			}

			cfg, err := config.Load()
			if err != nil {
				return err
			}

			inst, err := cfg.Find(name)
			if err != nil {
				return err
			}

			if inst.Storage == to {
				return fmt.Errorf("instance %q already uses %q storage", name, to)
			}

			if err := credentials.MigrateStorage(name, inst.Storage, to); err != nil {
				return fmt.Errorf("migrate storage: %w", err)
			}

			inst.Storage = to
			if err := cfg.Save(); err != nil {
				return err
			}

			fprint(cmd.OutOrStdout(), "Migrated instance %q storage to %q\n", name, to)
			return nil
		},
	}

	cmd.Flags().String("to", "", "Target storage mode: plaintext or encrypted (required)")
	_ = cmd.MarkFlagRequired("to")

	return cmd
}

func newInstanceUpdateTokenCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "update-token <name>",
		Short: "Update the Rancher API token for an instance",
		Long:  "Set a new Rancher API token after the current one becomes invalid or expired.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			cfg, err := config.Load()
			if err != nil {
				return err
			}
			inst, err := cfg.Find(name)
			if err != nil {
				return err
			}

			if tokenPageURL, err := rancher.TokenCreatePageURL(inst.URL); err != nil {
				return fmt.Errorf("build token page URL: %w", err)
			} else {
				prompt.PrintTokenCreateHint(cmd.OutOrStdout(), tokenPageURL)
			}

			token, err := prompt.ReadSecret(os.Stdin, cmd.OutOrStdout(), "New Rancher API token")
			if err != nil {
				return err
			}
			if token == "" {
				return fmt.Errorf("token must not be empty")
			}

			client, err := rancher.NewClient(inst.URL, token, inst.InsecureSkipVerify)
			if err != nil {
				return err
			}
			if err := client.ValidateToken(context.Background()); err != nil {
				return handleRancherError(name, err)
			}

			store, err := credentials.NewStore(inst.Storage)
			if err != nil {
				return err
			}
			if err := store.Set(name, token); err != nil {
				return err
			}

			fprint(cmd.OutOrStdout(), "Updated token for instance %q\n", name)
			return nil
		},
	}
}
