#!/usr/bin/env bash
# Acceptance check I8.1-a1: a downloaded release binary and a go install build convert the
# fixture identically. Usage: hack/verify-release.sh PATH_TO_BINARY [vX.Y.Z]
# With a version, go install fetches that tag from the module proxy (needs network);
# without one, it installs the working tree.
set -euo pipefail
cd "$(dirname "$0")/.."
binary="$(cd "$(dirname "${1:?usage: hack/verify-release.sh BINARY [vX.Y.Z]}")" && pwd)/$(basename "$1")"
version="${2:-}"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
gobin="$work/gobin"
if [ -n "$version" ]; then
  GOBIN="$gobin" go install "github.com/longhand-io/scrivener-to-obsidian/cmd/scriv2obsidian@$version"
else
  GOBIN="$gobin" go install ./cmd/scriv2obsidian
fi
go run ./hack/mkfixture "$work" >/dev/null
export SOURCE_DATE_EPOCH=0
"$binary" convert -out "$work/downloaded" "$work/Fixture Novel.scriv" >/dev/null
"$gobin/scriv2obsidian" convert -out "$work/installed" "$work/Fixture Novel.scriv" >/dev/null
echo "downloaded: $("$binary" version)"
echo "installed:  $("$gobin/scriv2obsidian" version)"
if diff -r "$work/downloaded" "$work/installed"; then
  echo "identical output on the fixture: $(find "$work/downloaded" -type f | wc -l | tr -d ' ') files"
else
  echo "output differs" >&2
  exit 1
fi
