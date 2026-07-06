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
	return mergeKubeconfigWithName(instance, clusterName, content, "")
}

func mergeKubeconfigWithName(instance, clusterName, content, contextName string) error {
	prefix := mergeContextName(instance, clusterName, "", contextName)

	incoming, err := clientcmd.Load([]byte(content))
	if err != nil {
		return fmt.Errorf("parse kubeconfig to merge: %w", err)
	}
	normalizeIncoming(incoming, prefix)

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

// contextExists reports whether a context name is already present in ~/.kube/config.
func contextExists(contextName string) (bool, string, error) {
	configPath, err := kubeconfigPath()
	if err != nil {
		return false, "", err
	}

	cfg, err := loadKubeConfig(configPath)
	if err != nil {
		return false, "", err
	}
	_, ok := cfg.Contexts[contextName]
	return ok, configPath, nil
}

type mergePromptOptions struct {
	Merge       bool
	Replace     bool
	Prefix      string
	ContextName string
	In          io.Reader
	Out         io.Writer
	IsTTY       bool
}

func offerMergeKubeconfig(opts mergePromptOptions, instance, clusterName, content string) error {
	prefix := mergeContextName(instance, clusterName, opts.Prefix, opts.ContextName)

	configPath, err := kubeconfigPath()
	if err != nil {
		return err
	}

	doMerge := opts.Merge
	if !doMerge {
		if !opts.IsTTY {
			return nil
		}
		question := fmt.Sprintf("Add context %q to %s", prefix, configPath)
		doMerge, err = prompt.Confirm(opts.In, opts.Out, question, false)
		if err != nil {
			return err
		}
	}
	if !doMerge {
		return nil
	}

	exists, _, err := contextExists(prefix)
	if err != nil {
		return err
	}

	if exists && !opts.Replace {
		if !opts.IsTTY {
			return fmt.Errorf("context %q already exists in %s; use --replace to overwrite", prefix, configPath)
		}
		question := fmt.Sprintf("Context %q already exists in %s. Replace it", prefix, configPath)
		replace, err := prompt.Confirm(opts.In, opts.Out, question, false)
		if err != nil {
			return err
		}
		if !replace {
			fprintln(opts.Out, "Skipped merge into kubeconfig.")
			return nil
		}
	}

	if err := mergeKubeconfigWithName(instance, clusterName, content, prefix); err != nil {
		return err
	}
	fprint(opts.Out, "Merged context %q into %s\n", prefix, configPath)
	return nil
}

func mergePrefix(instance, clusterName string) string {
	return instance + "-" + clusterName
}

func mergeContextName(instance, clusterName, prefix, contextName string) string {
	contextName = strings.TrimSpace(contextName)
	if contextName != "" {
		return contextName
	}

	prefix = strings.TrimSpace(prefix)
	if prefix == "" {
		return mergePrefix(instance, clusterName)
	}
	return prefix + "-" + clusterName
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
