# Rancher instances

The `rancher-instance` command group manages your Rancher connections — URLs, tokens, and storage preferences.

## Add an instance

```bash
kubectl sheep rancher-instance add <name> <url> [flags]
```

| Flag | Description |
|------|-------------|
| `--storage` | `encrypted` (default) or `plaintext` |
| `--insecure` | Skip TLS certificate verification |
| `--open` | Open the Rancher API key page in the browser |

Positional URL (no `--url` flag):

```bash
kubectl sheep rancher-instance add prod https://rancher.example.com
```

Interactive wizard when run with no arguments on a TTY:

```bash
kubectl sheep rancher-instance add
```

## List instances

```bash
kubectl sheep rancher-instance list
```

```
NAME   URL                          STORAGE    INSECURE
prod   https://rancher.example.com  encrypted  false
```

## Update token

When a Rancher API token expires:

```bash
kubectl sheep rancher-instance update-token prod
kubectl sheep rancher-instance update-token prod --open
```

Interactive — pick the instance:

```bash
kubectl sheep rancher-instance update-token
```

## Change storage mode

Migrate a token between plaintext and encrypted backends:

```bash
kubectl sheep rancher-instance set-storage prod --to=plaintext
kubectl sheep rancher-instance set-storage prod --to=encrypted
```

## List remote clusters

Inventory on the Rancher server (not local kubeconfigs):

```bash
kubectl sheep rancher-instance clusters list prod
```

Interactive — pick the instance:

```bash
kubectl sheep rancher-instance clusters list
```

## Remove an instance

Deletes the instance config and stored credentials:

```bash
kubectl sheep rancher-instance remove prod
```

Interactive — pick the instance:

```bash
kubectl sheep rancher-instance remove
```

## Configuration file

Instance metadata is stored in `~/.config/kubectl-sheep/instances.yaml`. See [Configuration](configuration.md) for all paths.
