# Install

kubectl-sheep is a [Krew](https://krew.sigs.k8s.io/) plugin:

```bash
kubectl krew install sheep
```

If you do not have Krew yet, follow the [Krew install guide](https://krew.sigs.k8s.io/docs/user-guide/setup/install/).

## Verify

```bash
kubectl sheep version
kubectl sheep --help
```

## Upgrade

```bash
kubectl krew upgrade sheep
```

## Shell completion

kubectl plugins are typically completed via kubectl itself:

```bash
kubectl completion bash  # or zsh
```

## Build from source

For development or contributing:

```bash
git clone https://github.com/ahoz/kubectl-sheep.git
cd kubectl-sheep
go build -o kubectl-sheep ./cmd/kubectl-sheep
chmod +x kubectl-sheep
mv kubectl-sheep "$(dirname "$(which kubectl)")/"
```

Requirements: Go 1.26+ (see `go.mod`).
