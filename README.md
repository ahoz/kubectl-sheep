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
# Manage Rancher instance connections
kubectl sheep rancher-instance add prod https://rancher.example.com --storage=encrypted
kubectl sheep rancher-instance add prod https://rancher.example.com --open
# Interactive: kubectl sheep rancher-instance add
# Interactive: kubectl sheep rancher-instance add prod
kubectl sheep rancher-instance add prod https://rancher.example.com --auth-login --auth-username alice
kubectl sheep rancher-instance list
kubectl sheep rancher-instance set-storage prod --to=plaintext
kubectl sheep rancher-instance update-token prod --open
kubectl sheep rancher-instance update-token prod --auth-login --auth-username alice
kubectl sheep rancher-instance remove prod

# List clusters on a Rancher instance (remote inventory)
kubectl sheep rancher-instance clusters list prod

# Kubeconfigs (local artifacts)
# Interactive: kubectl sheep kubeconfig get
kubectl sheep kubeconfig list prod
kubectl sheep kubeconfig get prod my-cluster
kubectl sheep kubeconfig get prod c-m-abc123 --merge --prefix prod
kubectl sheep kubeconfig get prod c-m-abc123 --merge --context-name prod-dev
kubectl sheep kubeconfig fetch prod my-cluster
kubectl sheep kubeconfig fetch prod --all
kubectl sheep kubeconfig refresh prod my-cluster
kubectl sheep kubeconfig refresh prod --all
kubectl sheep kubeconfig install-exec prod c-m-abc123 --context-name prod-dev

# kubeconfig get interactively offers to merge into ~/.kube/config as <rancher-instance>-<cluster>
# Non-interactive merge / replace:
kubectl sheep kubeconfig get prod my-cluster --merge
kubectl sheep kubeconfig get prod my-cluster --merge --replace

# Optional: merge contexts into ~/.kube/config for bulk commands
kubectl sheep kubeconfig fetch prod --all --merge
kubectl sheep kubeconfig refresh prod --all --merge
```

### Interactive mode

When stdin is a TTY, several commands prompt for missing arguments instead of failing:

| Command | Prompts for |
|---------|-------------|
| `rancher-instance add` | name, URL, storage, TLS skip, open browser |
| `rancher-instance update-token` | instance name |
| `kubeconfig get` | instance, cluster |
| `kubeconfig fetch` | instance, one cluster or `--all` |
| `kubeconfig refresh` | instance, one stored cluster or `--all` |

Pass `--no-input` to disable prompts (for scripts and CI).

```bash
# Fully interactive
kubectl sheep rancher-instance add
kubectl sheep kubeconfig get
kubectl sheep kubeconfig fetch
kubectl sheep kubeconfig refresh

# Partial args — only missing values are prompted
kubectl sheep rancher-instance add prod
kubectl sheep kubeconfig get prod
```

## Authentication

By default, `rancher-instance add` and `rancher-instance update-token` print the Rancher API key
page and prompt for a Bearer Token. Use `--open` to open that page in the
default browser.

For Rancher auth providers that support API login, kubectl-sheep can create the
Rancher API token directly:

```bash
kubectl sheep rancher-instance add prod \
  https://rancher.example.com \
  --auth-login \
  --auth-username alice \
  --auth-provider-type activeDirectory \
  --auth-provider-id activeDirectory
```

`--auth-provider-type` and `--auth-provider-id` default to `activeDirectory`.
For OpenLDAP, you can either set them explicitly or use the LDAP shortcut:

```bash
kubectl sheep rancher-instance add prod \
  https://rancher.example.com \
  --ldap-login \
  --ldap-username alice
```

## Context names

`kubeconfig get` accepts either a Rancher cluster name or ID. Use this to merge
only the clusters you want:

```bash
kubectl sheep rancher-instance clusters list prod
kubectl sheep kubeconfig get prod c-m-abc123 --merge
```

Merged contexts are named `<rancher-instance>-<cluster>` by default. Use `--prefix` to
replace the rancher-instance prefix, or `--context-name` to set the exact context name:

```bash
kubectl sheep kubeconfig get prod c-m-abc123 --merge --prefix team-a
kubectl sheep kubeconfig get prod c-m-abc123 --merge --context-name prod-dev
```

## Exec kubeconfigs

Use `kubeconfig install-exec` to create a stable kubeconfig context without storing
Rancher-generated Kubernetes credentials in `~/.kube/config`. The merged user
entry uses Kubernetes' exec credential plugin support and calls `kubectl-sheep`
whenever kubectl needs credentials.

```bash
kubectl sheep kubeconfig install-exec prod c-m-abc123 --context-name prod-dev
kubectl --context prod-dev get pods
```

The generated kubeconfig user looks like this:

```yaml
users:
- name: prod-dev
  user:
    exec:
      apiVersion: client.authentication.k8s.io/v1
      command: kubectl-sheep
      interactiveMode: IfAvailable
      args:
      - auth
      - exec
      - prod
      - c-m-abc123
```

Each user still needs to configure their local Rancher instance once with
`rancher-instance add` or `rancher-instance update-token`. After that, the same exec-based
kubeconfig can be shared without embedding tokens.

For non-interactive kubectl calls, make sure the local Rancher token can be read
without a passphrase prompt. For testing, or for environments where the
passphrase-protected file backend is not suitable for exec plugins, use
plaintext storage intentionally:

```bash
kubectl sheep rancher-instance add prod \
  https://rancher.example.com \
  --auth-login \
  --auth-username alice \
  --storage plaintext
```

The plaintext token is stored in `~/.config/kubectl-sheep/credentials.plain.yaml`
with file mode `0600`; do not commit or share that file.

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
