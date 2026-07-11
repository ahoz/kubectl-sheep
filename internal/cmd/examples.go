package cmd

// Example blocks for cobra help output. Prefix with "kubectl sheep" as users invoke
// the plugin through kubectl.

const (
	exRoot = `  # Register a Rancher instance
  kubectl sheep rancher-instance add prod https://rancher.example.com

  # Fetch a kubeconfig and merge it into ~/.kube/config
  kubectl sheep kubeconfig get prod my-cluster --merge

  # Interactive mode — omit arguments on a TTY
  kubectl sheep rancher-instance add
  kubectl sheep kubeconfig get`

	exVersion = `  kubectl sheep version`

	exRancherInstance = `  kubectl sheep rancher-instance add prod https://rancher.example.com
  kubectl sheep rancher-instance list
  kubectl sheep rancher-instance clusters list prod`

	exRancherInstanceAdd = `  # Register with encrypted token storage
  kubectl sheep rancher-instance add prod https://rancher.example.com --storage=encrypted

  # Interactive wizard (prompts for name, URL, and options)
  kubectl sheep rancher-instance add

  # Log in via Rancher auth provider instead of pasting a token
  kubectl sheep rancher-instance add prod https://rancher.example.com \
    --auth-login --auth-username alice`

	exRancherInstanceList = `  kubectl sheep rancher-instance list`

	exRancherInstanceRemove = `  kubectl sheep rancher-instance remove prod`

	exRancherInstanceSetStorage = `  kubectl sheep rancher-instance set-storage prod --to=encrypted`

	exRancherInstanceUpdateToken = `  kubectl sheep rancher-instance update-token prod
  kubectl sheep rancher-instance update-token prod --open

  # Interactive — pick the instance
  kubectl sheep rancher-instance update-token`

	exRancherInstanceClusters = `  kubectl sheep rancher-instance clusters list prod`

	exRancherInstanceClustersList = `  kubectl sheep rancher-instance clusters list prod`

	exKubeconfig = `  kubectl sheep kubeconfig list prod
  kubectl sheep kubeconfig get prod my-cluster
  kubectl sheep kubeconfig fetch prod --all`

	exKubeconfigList = `  kubectl sheep kubeconfig list prod`

	exKubeconfigGet = `  # Fetch and store locally
  kubectl sheep kubeconfig get prod my-cluster

  # Fetch and merge into ~/.kube/config
  kubectl sheep kubeconfig get prod my-cluster --merge

  # Interactive — pick instance and cluster
  kubectl sheep kubeconfig get`

	exKubeconfigFetch = `  # Fetch one cluster
  kubectl sheep kubeconfig fetch prod my-cluster

  # Fetch every cluster on the instance
  kubectl sheep kubeconfig fetch prod --all

  # Interactive — pick instance and scope
  kubectl sheep kubeconfig fetch`

	exKubeconfigRefresh = `  # Refresh one stored kubeconfig
  kubectl sheep kubeconfig refresh prod my-cluster

  # Refresh everything stored for the instance
  kubectl sheep kubeconfig refresh prod --all

  # Interactive — pick instance and scope
  kubectl sheep kubeconfig refresh`

	exKubeconfigInstallExec = `  kubectl sheep kubeconfig install-exec prod c-m-abc123
  kubectl sheep kubeconfig install-exec prod c-m-abc123 --context-name prod-dev`
)
