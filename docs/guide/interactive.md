# Interactive mode

When stdin is a TTY, kubectl-sheep prompts for missing arguments instead of failing. Pass `--no-input` to disable prompts (scripts, CI).

Selection menus use arrow-key navigation, fuzzy search (`/`), and a 🐑 pointer on the active row. Free-text fields (name, URL, tokens) use simple line prompts with a trailing colon.

## Commands with interactive support

| Command | Prompts for |
|---------|-------------|
| `rancher-instance add` | Name, URL, storage, TLS skip, open browser, token |
| `rancher-instance remove` | Instance |
| `rancher-instance update-token` | Instance, token |
| `rancher-instance clusters list` | Instance |
| `kubeconfig list` | Instance |
| `kubeconfig get` | Instance, scope (one / multiple / all), cluster(s) |
| `kubeconfig refresh` | Instance, one stored cluster or all |

## Examples

Fully interactive:

```bash
kubectl sheep rancher-instance add
kubectl sheep rancher-instance remove
kubectl sheep rancher-instance clusters list
kubectl sheep kubeconfig list
kubectl sheep kubeconfig get
kubectl sheep kubeconfig refresh
```

Partial arguments — only missing values are prompted:

```bash
kubectl sheep rancher-instance add prod          # prompts for URL
kubectl sheep kubeconfig get prod              # prompts for scope and cluster(s)
```

## What the menus look like

List selections (instances, clusters, scope):

```
Use the arrow keys to navigate: ↓ ↑ → ← and / toggles search
Rancher instance
  🐑 paas
   dev

--------- Info ----------
Name: paas
URL: https://paas.example.com
```

Multi-cluster fetch uses **space** to toggle `[x]` on clusters and **enter** to confirm.

## After fetch

Kubeconfigs are always saved and merged. You will see a summary like:

```
✓ Saved and merged kubeconfig for "production"
  ~/.kube/sheep/prod/c-m-abc123.yaml
  Context "prod-production" → ~/.kube/config
```

## Disable prompts

```bash
kubectl sheep --no-input kubeconfig get prod
# Error: specify a cluster or pass --all
```

Use explicit arguments in automation:

```bash
kubectl sheep kubeconfig get prod my-cluster
kubectl sheep kubeconfig get prod --all
```

## Encrypted storage

The passphrase for encrypted token storage is requested once per process, even when multiple Rancher API calls happen during one interactive session. Passphrase input does not echo characters to the terminal.

## Shell completion

Enable tab completion separately — see [Install](install.md#shell-completion).
