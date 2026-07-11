# Quick start

This walkthrough registers a Rancher instance, fetches one cluster kubeconfig, and merges it into `~/.kube/config`.

## 1. Add a Rancher instance

```bash
kubectl sheep rancher-instance add prod https://rancher.example.com --storage=encrypted
```

kubectl-sheep prints the Rancher API key page URL. Create a key in the UI and paste the Bearer token when prompted.

Use `--open` to open the token page in your browser:

```bash
kubectl sheep rancher-instance add prod https://rancher.example.com --open
```

Or run fully interactive:

```bash
kubectl sheep rancher-instance add
```

## 2. List remote clusters

```bash
kubectl sheep rancher-instance clusters list prod
```

```
ID           NAME          STATE
c-m-abc123   production    active
c-m-def456   staging       active
```

## 3. Fetch a kubeconfig

Non-interactive:

```bash
kubectl sheep kubeconfig get prod production
```

Interactive (pick instance and cluster):

```bash
kubectl sheep kubeconfig get
```

The kubeconfig is saved to `~/.kube/sheep/prod/<cluster-id>.yaml`.

## 4. Merge into ~/.kube/config

On a TTY, `kubeconfig get` offers to merge after saving. You can also force it:

```bash
kubectl sheep kubeconfig get prod production --merge
```

Default context name: `prod-production` (`<instance>-<cluster>`).

## 5. Use kubectl

```bash
kubectl --context prod-production get nodes
```

## Bulk fetch (optional)

Download every cluster on an instance:

```bash
kubectl sheep kubeconfig fetch prod --all
```


Merge all fetched contexts in one go:

```bash
kubectl sheep kubeconfig fetch prod --all --merge
```

## What's next?

- [Rancher instances](rancher-instances.md) — token rotation, storage migration, removal
- [Interactive mode](interactive.md) — prompts and `--no-input`
- [Exec contexts](exec-kubeconfigs.md) — share kubeconfigs without embedded tokens
