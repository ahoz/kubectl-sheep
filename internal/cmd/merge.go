package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ahoz/kubectl-sheep/internal/prompt"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

func mergeKubeconfig(instance, clusterName, content string) error {
	contextName := mergeContextName(instance, clusterName)

	incoming, err := clientcmd.Load([]byte(content))
	if err != nil {
		return fmt.Errorf("parse kubeconfig to merge: %w", err)
	}
	normalizeIncoming(incoming, contextName)

	configPath, err := kubeconfigPath()
	if err != nil {
		return err
	}

	dest, err := loadKubeConfig(configPath)
	if err != nil {
		return err
	}

	mergeConfigs(dest, incoming)

	if err := os.MkdirAll(filepath.Dir(configPath), 0o700); err != nil {
		return fmt.Errorf("create kube config directory: %w", err)
	}
	if err := clientcmd.WriteToFile(*dest, configPath); err != nil {
		return fmt.Errorf("write %s: %w", configPath, err)
	}
	return nil
}

func loadKubeConfig(configPath string) (*clientcmdapi.Config, error) {
	_, err := os.Stat(configPath)
	if err == nil {
		cfg, err := clientcmd.LoadFromFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("load %s: %w", configPath, err)
		}
		return cfg, nil
	}
	if os.IsNotExist(err) {
		return clientcmdapi.NewConfig(), nil
	}
	return nil, fmt.Errorf("stat %s: %w", configPath, err)
}

func reportKubeconfigSaved(out io.Writer, interactive bool, clusterName, path, contextName, configPath string) {
	if interactive {
		prompt.Success(out, fmt.Sprintf(`Saved and merged kubeconfig for %q`, clusterName))
		prompt.Note(out, path)
		prompt.Note(out, fmt.Sprintf(`Context %q → %s`, contextName, configPath))
		return
	}
	fprint(out, "Saved kubeconfig for %q to %s, merged context %q into %s\n", clusterName, path, contextName, configPath)
}

func mergeContextName(instance, clusterName string) string {
	return instance + "-" + clusterName
}

type execKubeconfigOptions struct {
	contextName string
	command     string
	args        []string
}

func buildExecKubeconfig(content string, opts execKubeconfigOptions) (string, error) {
	contextName := strings.TrimSpace(opts.contextName)
	if contextName == "" {
		return "", fmt.Errorf("context name must not be empty")
	}
	command := strings.TrimSpace(opts.command)
	if command == "" {
		return "", fmt.Errorf("exec command must not be empty")
	}

	incoming, err := clientcmd.Load([]byte(content))
	if err != nil {
		return "", fmt.Errorf("parse kubeconfig for exec install: %w", err)
	}

	srcName := incoming.CurrentContext
	if srcName == "" {
		for name := range incoming.Contexts {
			srcName = name
			break
		}
	}
	if srcName == "" {
		return "", fmt.Errorf("kubeconfig does not contain a context")
	}

	src, ok := incoming.Contexts[srcName]
	if !ok {
		return "", fmt.Errorf("kubeconfig context %q not found", srcName)
	}

	cluster, ok := incoming.Clusters[src.Cluster]
	if !ok {
		return "", fmt.Errorf("kubeconfig cluster %q not found", src.Cluster)
	}

	out := clientcmdapi.NewConfig()
	out.Clusters[contextName] = cluster
	out.AuthInfos[contextName] = &clientcmdapi.AuthInfo{
		Exec: &clientcmdapi.ExecConfig{
			APIVersion:      "client.authentication.k8s.io/v1",
			Command:         command,
			Args:            opts.args,
			InteractiveMode: clientcmdapi.IfAvailableExecInteractiveMode,
		},
	}
	out.Contexts[contextName] = &clientcmdapi.Context{
		Cluster:    contextName,
		AuthInfo:   contextName,
		Namespace:  src.Namespace,
		Extensions: src.Extensions,
	}
	out.CurrentContext = contextName

	data, err := clientcmd.Write(*out)
	if err != nil {
		return "", fmt.Errorf("write exec kubeconfig: %w", err)
	}
	return string(data), nil
}

func kubeconfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}
	if env := os.Getenv("KUBECONFIG"); env != "" {
		parts := strings.Split(env, string(os.PathListSeparator))
		if len(parts) > 0 && parts[0] != "" {
			return parts[0], nil
		}
	}
	return filepath.Join(home, ".kube", "config"), nil
}

func normalizeIncoming(cfg *clientcmdapi.Config, prefix string) {
	srcName := cfg.CurrentContext
	if srcName == "" {
		for name := range cfg.Contexts {
			srcName = name
			break
		}
	}

	src, ok := cfg.Contexts[srcName]
	if !ok {
		return
	}

	cluster, ok := cfg.Clusters[src.Cluster]
	if !ok {
		return
	}
	auth, ok := cfg.AuthInfos[src.AuthInfo]
	if !ok {
		return
	}

	normalized := clientcmdapi.NewConfig()
	normalized.Clusters[prefix] = cluster
	normalized.AuthInfos[prefix] = auth
	normalized.Contexts[prefix] = &clientcmdapi.Context{
		Cluster:    prefix,
		AuthInfo:   prefix,
		Namespace:  src.Namespace,
		Extensions: src.Extensions,
	}
	normalized.CurrentContext = prefix

	*cfg = *normalized
}

func mergeConfigs(dest, incoming *clientcmdapi.Config) {
	for name, cluster := range incoming.Clusters {
		dest.Clusters[name] = cluster
	}
	for name, auth := range incoming.AuthInfos {
		dest.AuthInfos[name] = auth
	}
	for name, ctx := range incoming.Contexts {
		dest.Contexts[name] = ctx
	}

	if dest.CurrentContext == "" && incoming.CurrentContext != "" {
		dest.CurrentContext = incoming.CurrentContext
	}
}
