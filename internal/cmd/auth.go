package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ahoz/kubectl-sheep/internal/instance"
	"github.com/ahoz/kubectl-sheep/internal/rancher"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientauthv1 "k8s.io/client-go/pkg/apis/clientauthentication/v1"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

func newAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Authentication helpers for kubeconfig exec users",
		Long:  "Internal commands used by exec-based kubeconfig contexts installed via kubeconfig install-exec.",
	}

	cmd.AddCommand(newAuthExecCmd())

	return cmd
}

func newAuthExecCmd() *cobra.Command {
	return &cobra.Command{
		Use:    "exec <instance> <cluster>",
		Short:  "Return Kubernetes ExecCredential for a Rancher cluster",
		Long:   "Fetch a Rancher-generated kubeconfig and return its credentials as Kubernetes ExecCredential JSON.",
		Hidden: true,
		Args:   cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			instanceName := args[0]
			clusterRef := args[1]

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

			credential, err := execCredentialFromKubeconfig(content)
			if err != nil {
				return err
			}

			encoder := json.NewEncoder(cmd.OutOrStdout())
			return encoder.Encode(credential)
		},
	}
}

func execCredentialFromKubeconfig(content string) (*clientauthv1.ExecCredential, error) {
	cfg, err := clientcmd.Load([]byte(content))
	if err != nil {
		return nil, fmt.Errorf("parse kubeconfig for exec credential: %w", err)
	}

	auth, err := currentAuthInfo(cfg)
	if err != nil {
		return nil, err
	}

	status := &clientauthv1.ExecCredentialStatus{}
	if token := strings.TrimSpace(auth.Token); token != "" {
		status.Token = token
	}
	if len(auth.ClientCertificateData) > 0 && len(auth.ClientKeyData) > 0 {
		status.ClientCertificateData = string(auth.ClientCertificateData)
		status.ClientKeyData = string(auth.ClientKeyData)
	}
	if status.Token == "" && status.ClientCertificateData == "" {
		return nil, fmt.Errorf("rancher kubeconfig does not contain token or client certificate credentials")
	}

	return &clientauthv1.ExecCredential{
		TypeMeta: metav1Type("ExecCredential", "client.authentication.k8s.io/v1"),
		Status:   status,
	}, nil
}

func currentAuthInfo(cfg *clientcmdapi.Config) (*clientcmdapi.AuthInfo, error) {
	contextName := cfg.CurrentContext
	if contextName == "" {
		for name := range cfg.Contexts {
			contextName = name
			break
		}
	}
	if contextName == "" {
		return nil, fmt.Errorf("kubeconfig does not contain a context")
	}

	ctx, ok := cfg.Contexts[contextName]
	if !ok {
		return nil, fmt.Errorf("kubeconfig current context %q not found", contextName)
	}

	auth, ok := cfg.AuthInfos[ctx.AuthInfo]
	if !ok {
		return nil, fmt.Errorf("kubeconfig auth info %q not found", ctx.AuthInfo)
	}
	return auth, nil
}

func metav1Type(kind, apiVersion string) metav1.TypeMeta {
	return metav1.TypeMeta{
		Kind:       kind,
		APIVersion: apiVersion,
	}
}
