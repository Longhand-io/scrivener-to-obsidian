#!/usr/bin/env bash
# Print the CHANGELOG section for one version. That text is the GitHub release notes.
# Usage: hack/release-notes.sh vX.Y.Z
# Fails when the heading is still "## Unreleased: vX.Y.Z"; rename it to "## vX.Y.Z" when cutting.
set -euo pipefail
cd "$(dirname "$0")/.."
version="${1:?usage: hack/release-notes.sh vX.Y.Z}"
notes="$(awk -v v="$version" '
  /^## / { if (found) exit; found = ($2 == v); next }
  found  { lines[n++] = $0 }
  END {
    while (n > 0 && lines[n-1] == "") n--
    start = 0
    while (start < n && lines[start] == "") start++
    for (i = start; i < n; i++) print lines[i]
  }' CHANGELOG.md)"
if [ -z "$notes" ]; then
  echo "CHANGELOG.md has no \"## $version\" section. Rename the Unreleased heading before tagging." >&2
  exit 1
fi
printf '%s\n' "$notes"
