# Working with kubeconfigs

The `kubeconfig` command group downloads Rancher-generated kubeconfigs, caches them locally, and optionally merges them into `~/.kube/config`.

## List local kubeconfigs

```bash
kubectl sheep kubeconfig list prod
```

```
CLUSTER ID   CLUSTER NAME   FETCHED AT           PATH
c-m-abc123   production     2026-07-05 14:30:00  ~/.kube/sheep/prod/c-m-abc123.yaml
```

## Get one cluster

Download and store a single kubeconfig:

```bash
kubectl sheep kubeconfig get prod my-cluster
kubectl sheep kubeconfig get prod c-m-abc123
```

Accepts cluster **name** or **ID**.

### Merge flags

| Flag | Description |
|------|-------------|
| `--merge` | Merge into `~/.kube/config` without prompting |
| `--replace` | Overwrite an existing context (with `--merge`) |
| `--prefix` | Context name prefix (default: instance name) |
| `--context-name` | Exact context name |

```bash
kubectl sheep kubeconfig get prod c-m-abc123 --merge
kubectl sheep kubeconfig get prod c-m-abc123 --merge --context-name prod-dev
kubectl sheep kubeconfig get prod c-m-abc123 --merge --prefix team-a
```

Default context name: `<instance>-<cluster>` (e.g. `prod-production`).

## Fetch

Alias for bulk download workflows. Same underlying fetch as `get`:

```bash
# One cluster
kubectl sheep kubeconfig fetch prod my-cluster

# All clusters on the instance
kubectl sheep kubeconfig fetch prod --all

# Fetch and merge everything
kubectl sheep kubeconfig fetch prod --all --merge
```


## Refresh

Re-download kubeconfigs that are already stored locally:

```bash
kubectl sheep kubeconfig refresh prod my-cluster
kubectl sheep kubeconfig refresh prod --all
kubectl sheep kubeconfig refresh prod --all --merge
```

Refresh reports whether each kubeconfig changed and prints token expiry hints when available.


## Storage layout

| Path | Content |
|------|---------|
| `~/.kube/sheep/<instance>/<cluster-id>.yaml` | Kubeconfig content |
| `~/.kube/sheep/<instance>/<cluster-id>.meta.yaml` | Cluster name, fetch timestamp |

See [Configuration](configuration.md) for the full path reference.

## Context naming cheat sheet

| Command | Resulting context |
|---------|-------------------|
| `get prod my-cluster` (interactive merge) | `prod-my-cluster` |
| `--prefix team-a` | `team-a-my-cluster` |
| `--context-name prod-dev` | `prod-dev` |
