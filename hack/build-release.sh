#!/usr/bin/env bash
# Build the release archives into dist/: one per target, plus a sha256 checksum file.
# Usage: hack/build-release.sh vX.Y.Z
# Static binaries, no cgo, paths trimmed, version stamped from the argument.
set -euo pipefail
cd "$(dirname "$0")/.."
version="${1:?usage: hack/build-release.sh vX.Y.Z}"
targets="darwin/arm64 darwin/amd64 linux/amd64 linux/arm64 windows/amd64"
rm -rf dist
mkdir -p dist
for target in $targets; do
  goos="${target%/*}"
  goarch="${target#*/}"
  name="scriv2obsidian_${version}_${goos}_${goarch}"
  bin="scriv2obsidian"
  [ "$goos" = windows ] && bin="scriv2obsidian.exe"
  mkdir -p "dist/$name"
  CGO_ENABLED=0 GOOS="$goos" GOARCH="$goarch" \
    go build -trimpath -ldflags "-s -w -X main.version=${version}" \
    -o "dist/$name/$bin" ./cmd/scriv2obsidian
  cp LICENSE NOTICE README.md "dist/$name/"
  if [ "$goos" = windows ]; then
    (cd dist && zip -qr "$name.zip" "$name")
  else
    tar -C dist -czf "dist/$name.tar.gz" "$name"
  fi
  rm -r "dist/$name"
  echo "built $name"
done
if command -v sha256sum >/dev/null 2>&1; then sum=sha256sum; else sum="shasum -a 256"; fi
(cd dist && $sum ./*.tar.gz ./*.zip | sed 's#^\([0-9a-f]*\)  \./#\1  #' > "scriv2obsidian_${version}_checksums.txt")
cat "dist/scriv2obsidian_${version}_checksums.txt"
