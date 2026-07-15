# Exec kubeconfig contexts

`kubeconfig install-exec` creates a kubeconfig context that loads credentials on demand via Kubernetes' exec credential plugin — no long-lived token embedded in `~/.kube/config`.

## Install

```bash
kubectl sheep kubeconfig install-exec prod c-m-abc123
```

The context is saved as `prod-<cluster-name>` (same naming as a normal `kubeconfig get`) and merged into `~/.kube/config`.

| Flag | Description |
|------|-------------|
| `--exec-command` | Binary invoked by kubeconfig exec (default: `kubectl-sheep`) |

## Generated user entry

```yaml
users:
- name: prod-production
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

When you run `kubectl --context prod-production get pods`, kubectl calls `kubectl-sheep auth exec prod c-m-abc123`, which fetches a fresh kubeconfig from Rancher and returns an `ExecCredential`.

## Prerequisites

Each user must register the Rancher instance locally first:

```bash
kubectl sheep rancher-instance add prod https://rancher.example.com
```

The shared kubeconfig context does **not** contain the Rancher token — only the exec reference.

## Non-interactive / CI usage

Exec contexts need the Rancher token readable without a passphrase prompt. Use plaintext storage for that instance:

```bash
kubectl sheep rancher-instance add prod \
  https://rancher.example.com \
  --auth-login \
  --auth-username alice \
  --storage plaintext
```

> Plaintext tokens live in `~/.config/kubectl-sheep/credentials.plain.yaml` with mode `0600`. Never commit or share this file.

## Verify

```bash
kubectl --context prod-production get pods
kubectl config view --minify --context prod-production
```
