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
kubectl sheep instance add prod --url=https://rancher.example.com --open
kubectl sheep instance add prod --url=https://rancher.example.com --auth-login --auth-username alice
kubectl sheep instance list
kubectl sheep instance set-storage prod --to=plaintext
kubectl sheep instance update-token prod --open
kubectl sheep instance update-token prod --auth-login --auth-username alice
kubectl sheep instance remove prod

# Clusters
kubectl sheep cluster list prod
kubectl sheep cluster get prod my-cluster
kubectl sheep cluster get prod c-m-abc123 --merge --prefix prod
kubectl sheep cluster get prod c-m-abc123 --merge --context-name prod-dev
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

## Authentication

By default, `instance add` and `instance update-token` print the Rancher API key
page and prompt for a Bearer Token. Use `--open` to open that page in the
default browser.

For Rancher auth providers that support API login, kubectl-sheep can create the
Rancher API token directly:

```bash
kubectl sheep instance add dev-rancher \
  --url=https://rancher.example.com \
  --auth-login \
  --auth-username alice \
  --auth-provider-type activeDirectory \
  --auth-provider-id activeDirectory
```

`--auth-provider-type` and `--auth-provider-id` default to `activeDirectory`.
For OpenLDAP, you can either set them explicitly or use the LDAP shortcut:

```bash
kubectl sheep instance add prod \
  --url=https://rancher.example.com \
  --ldap-login \
  --ldap-username alice
```

## Context names

`cluster get` accepts either a Rancher cluster name or ID. Use this to merge
only the clusters you want:

```bash
kubectl sheep cluster list prod
kubectl sheep cluster get prod c-m-abc123 --merge
```

Merged contexts are named `<instance>-<cluster>` by default. Use `--prefix` to
replace the instance prefix, or `--context-name` to set the exact context name:

```bash
kubectl sheep cluster get prod c-m-abc123 --merge --prefix bv-dev
kubectl sheep cluster get prod c-m-abc123 --merge --context-name dev
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
