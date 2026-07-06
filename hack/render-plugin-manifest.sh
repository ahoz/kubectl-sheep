#!/usr/bin/env bash
# Renders plugin.yaml for a release with computed SHA256 checksums.
# Usage: render-plugin-manifest.sh <tag> [dist-dir]
# Example: render-plugin-manifest.sh v0.1.0 dist
set -euo pipefail

TAG="${1:?release tag required (e.g. v0.1.0)}"
DIST="${2:-dist}"
REPO="${GITHUB_REPOSITORY:-ahoz/kubectl-sheep}"

sha256_file() {
  local file="$1"
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$file" | awk '{print $1}'
  else
    shasum -a 256 "$file" | awk '{print $1}'
  fi
}

platform_block() {
  local os="$1"
  local arch="$2"
  local archive="${DIST}/kubectl-sheep_${os}_${arch}.tar.gz"

  if [[ ! -f "$archive" ]]; then
    echo "missing release archive: $archive" >&2
    exit 1
  fi

  local hash
  hash="$(sha256_file "$archive")"

  cat <<EOF
    - selector:
        matchLabels:
          os: ${os}
          arch: ${arch}
      uri: https://github.com/${REPO}/releases/download/${TAG}/kubectl-sheep_${os}_${arch}.tar.gz
      sha256: ${hash}
      bin: kubectl-sheep
EOF
}

cat <<EOF
apiVersion: krew.googlecontainertools.github.com/v1alpha2
kind: Plugin
metadata:
  name: sheep
spec:
  version: ${TAG}
  homepage: https://github.com/${REPO}
  shortDescription: Fetch and manage kubeconfigs from Rancher-managed clusters
  description: |
    A kubectl plugin to manage multiple Rancher instances, list their downstream
    clusters, and fetch/refresh kubeconfigs individually or in bulk. Rancher API
    tokens can be stored either as plaintext or encrypted (passphrase-protected file
    backend), selectable per instance.
  platforms:
EOF

platform_block linux amd64
platform_block linux arm64
platform_block darwin amd64
platform_block darwin arm64
