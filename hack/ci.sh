#!/usr/bin/env bash
# The gate. gofmt, vet, build, test with race. Standard library only, so no module downloads.
set -euo pipefail
cd "$(dirname "$0")/.."
unformatted="$(gofmt -l ./cmd ./internal 2>/dev/null || true)"
if [ -n "$unformatted" ]; then echo "gofmt needed:"; echo "$unformatted"; exit 1; fi
hack/verify-boilerplate.sh
go vet ./...
go build ./...
go test -race ./...
echo "gate green"
