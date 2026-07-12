#!/usr/bin/env bash
# Generates the custom release notes body (commit history) for GitHub releases.
# Intended to be prepended to GitHub's auto-generated notes (Full Changelog link).
#
# Usage: generate-release-notes-body.sh <tag>
# Example: generate-release-notes-body.sh v0.1.1
set -euo pipefail

TAG="${1:?release tag required (e.g. v0.1.1)}"

PREVIOUS_TAG="$(git tag --sort=-v:refname --merged "$TAG" | grep -Fxv "$TAG" | head -1 || true)"

github_username() {
  local email="$1"
  local name="$2"

  if [[ "$email" =~ ^[0-9]*\+([^@]+)@users\.noreply\.github\.com$ ]]; then
    echo "${BASH_REMATCH[1]}"
    return
  fi

  if [[ "$email" =~ ^([^@]+)@users\.noreply\.github\.com$ ]]; then
    echo "${BASH_REMATCH[1]}"
    return
  fi

  echo "$name"
}

echo "Commit History:"

if [[ -z "$PREVIOUS_TAG" ]]; then
  COMMIT_RANGE="$TAG"
else
  COMMIT_RANGE="${PREVIOUS_TAG}..${TAG}"
fi

while IFS=$'\t' read -r subject email name; do
  username="$(github_username "$email" "$name")"
  echo "- ${subject} [@${username}]"
done < <(git log --reverse --pretty=format:"%s%x09%ae%x09%an%n" "$COMMIT_RANGE")

echo ""
