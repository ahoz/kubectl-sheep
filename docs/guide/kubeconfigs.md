# Working with kubeconfigs

The `kubeconfig` command group downloads Rancher-generated kubeconfigs, caches them locally, and **automatically merges** them into `~/.kube/config`.

Merged contexts are always named `<rancher-instance>-<cluster-name>` (e.g. `prod-production`).

## List local kubeconfigs

```bash
kubectl sheep kubeconfig list prod
```

```
CLUSTER ID   CLUSTER NAME   FETCHED AT           PATH
c-m-abc123   production     2026-07-05 14:30:00  ~/.kube/sheep/prod/c-m-abc123.yaml
```

Interactive — pick the instance:

```bash
kubectl sheep kubeconfig list
```

## Get clusters

Download, store, and merge kubeconfigs from Rancher:

```bash
# One cluster (name or ID)
kubectl sheep kubeconfig get prod my-cluster
kubectl sheep kubeconfig get prod c-m-abc123

# All clusters on the instance
kubectl sheep kubeconfig get prod --all
```

Interactive — pick instance and scope (one, multiple, or all):

```bash
kubectl sheep kubeconfig get
```

After each fetch you will see where the kubeconfig was saved and which context was merged, for example:

```
✓ Saved and merged kubeconfig for "production"
  ~/.kube/sheep/prod/c-m-abc123.yaml
  Context "prod-production" → ~/.kube/config
```

## Refresh

Re-download kubeconfigs that are already stored locally (and re-merge into `~/.kube/config`):

```bash
kubectl sheep kubeconfig refresh prod my-cluster
kubectl sheep kubeconfig refresh prod --all
```

Refresh reports whether each kubeconfig changed and prints token expiry hints when available.

## Install exec context

Create a kubeconfig context that loads credentials on demand. See [Exec kubeconfig contexts](exec-kubeconfigs.md).

```bash
kubectl sheep kubeconfig install-exec prod c-m-abc123
```

Context name defaults to `prod-<cluster-name>`.

## Storage layout

| Path | Content |
|------|---------|
| `~/.kube/sheep/<instance>/<cluster-id>.yaml` | Kubeconfig content |
| `~/.kube/sheep/<instance>/<cluster-id>.meta.yaml` | Cluster name, fetch timestamp |
| `~/.kube/config` | Merged contexts (or first path in `$KUBECONFIG`) |

See [Configuration](configuration.md) for the full path reference.

## Context naming

| Scenario | Context name |
|----------|----------------|
| Instance `prod`, cluster `production` | `prod-production` |
| Instance `paas`, cluster `dev-gcp-01` | `paas-dev-gcp-01` |

Existing contexts with the same name are overwritten on re-fetch.
