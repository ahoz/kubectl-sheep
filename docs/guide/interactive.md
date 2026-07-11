# Interactive mode

When stdin is a TTY, kubectl-sheep prompts for missing arguments instead of failing. Pass `--no-input` to disable prompts (scripts, CI).

## Commands with interactive support

| Command | Prompts for |
|---------|-------------|
| `rancher-instance add` | Name, URL, storage, TLS skip, open browser |
| `rancher-instance update-token` | Instance name |
| `kubeconfig get` | Instance, cluster, merge confirmation |
| `kubeconfig fetch` | Instance, one cluster or all |
| `kubeconfig refresh` | Instance, one stored cluster or all |

## Examples

Fully interactive:

```bash
kubectl sheep rancher-instance add
kubectl sheep kubeconfig get
kubectl sheep kubeconfig fetch
kubectl sheep kubeconfig refresh
```

Partial arguments — only missing values are prompted:

```bash
kubectl sheep rancher-instance add prod          # prompts for URL
kubectl sheep kubeconfig get prod                # prompts for cluster
```

## What the prompts look like

Interactive flows use section dividers and numbered lists:

```
Fetch cluster kubeconfig

── Rancher instance ──

  1  main
     https://rancher.example.com/

  Choose [1-1] (or type a value): 1

── Cluster ──

  1  production
     c-m-abc123 · active

  Choose [1-1] (or type a value): 1

✓ Saved kubeconfig for "production"
  ~/.kube/sheep/main/c-m-abc123.yaml

── Merge into kubeconfig ──

  ~/.kube/config
  Add context "main-production"? [y/N]:
```

## Disable prompts

```bash
kubectl sheep --no-input kubeconfig get prod
# Error: cluster is required
```

Use explicit arguments and flags in automation:

```bash
kubectl sheep kubeconfig get prod my-cluster --merge --replace
```

## Encrypted storage

The passphrase for encrypted token storage is requested once per process, even when multiple Rancher API calls happen during one interactive session.
