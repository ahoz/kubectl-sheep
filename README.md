<p align="center">
  <img src="sheep.png" alt="kubectl-sheep logo" width="160">
</p>

# kubectl-sheep

A [Krew](https://krew.sigs.k8s.io/) plugin for managing kubeconfigs from Rancher-managed clusters.
## Install

> [!NOTE]
> This plugin has not been submitted to the [official Krew index](https://github.com/kubernetes-sigs/krew-index) yet. Install from source or a GitHub release until then.

```bash
kubectl krew install sheep
```

For local development, build and place the binary on your `PATH`:

```bash
go build -o kubectl-sheep ./cmd/kubectl-sheep
chmod +x kubectl-sheep
mv kubectl-sheep "$(dirname "$(which kubectl)")/"
```

## Usage

```bash
# Manage Rancher instances
kubectl sheep instance add prod --url=https://rancher.example.com --storage=encrypted
kubectl sheep instance list
kubectl sheep instance set-storage prod --to=plaintext
kubectl sheep instance update-token prod
kubectl sheep instance remove prod

# Clusters
kubectl sheep cluster list prod
kubectl sheep cluster get prod my-cluster
kubectl sheep cluster refresh prod my-cluster

# cluster get interactively offers to merge into ~/.kube/config as <instance>-<cluster>
# Non-interactive merge / replace:
kubectl sheep cluster get prod my-cluster --merge
kubectl sheep cluster get prod my-cluster --merge --replace

# Bulk operations
kubectl sheep fetch-all prod
kubectl sheep refresh-all prod

# Optional: merge contexts into ~/.kube/config for bulk commands
kubectl sheep fetch-all prod --merge
kubectl sheep refresh-all prod --merge
```

## Configuration

| Path | Purpose |
|------|---------|
| `~/.config/kubectl-sheep/instances.yaml` | Instance names, URLs, storage mode |
| `~/.config/kubectl-sheep/credentials.plain.yaml` | Plaintext tokens (`0600`) |
| `~/.config/kubectl-sheep/keys/` | Encrypted tokens (keyring FileBackend) |
| `~/.kube/sheep/<instance>/<cluster-id>.yaml` | Fetched kubeconfigs |
| `~/.kube/sheep/<instance>/<cluster-id>.meta.yaml` | Fetch metadata |

## Development

```bash
go test ./...
go build -o kubectl-sheep ./cmd/kubectl-sheep
```

## Credits

This project was herded 🐑 into existence with help from AI tooling 🪄

## License

Apache 2.0
