#!/bin/bash
# Builds release binaries for all supported platforms

set -euo pipefail
cd "$(dirname "$0")"

./make-agent.sh

go vet ./...

build() {
  local goos=$1 goarch=$2 name=$3
  local dir="build/${goos}_${goarch}"
  rm -rf "$dir"
  mkdir -p "$dir"
  echo "Building $dir"
  GOOS=$goos GOARCH=$goarch CGO_ENABLED=0 go build -o "$dir/jvm-mon"
  tar cvzf "build/jvm-mon-${name}.tgz" -C "$dir" jvm-mon
}

rm -f build/*.tgz
build darwin arm64 darwin-arm64
build darwin amd64 darwin-amd64
build linux amd64 linux-x64
