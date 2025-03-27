#!/bin/bash

set -euo pipefail

./make-agent.sh

export GOOS=darwin
export GOARCH=arm64
DIR="${GOOS}_${GOARCH}"
rm build/${DIR}/*
mkdir -p build/${DIR}
echo "Building $DIR"
go build -o build/${DIR}/jvm-mon

export GOOS=darwin
export GOARCH=amd64
DIR="${GOOS}_${GOARCH}"
rm build/${DIR}/*
mkdir -p build/${DIR}
echo "Building $DIR"
go build -o build/${DIR}/jvm-mon

export GOOS=linux
export GOARCH=amd64
DIR="${GOOS}_${GOARCH}"
rm build/${DIR}/*
mkdir -p build/${DIR}
echo "Building $DIR"
go build -o build/${DIR}/jvm-mon

rm build/*.tgz
tar cvzf build/jvm-mon-darwin-arm64.tgz -C build/darwin_arm64 jvm-mon
tar cvzf build/jvm-mon-darwin-amd64.tgz -C build/darwin_arm64 jvm-mon
tar cvzf build/jvm-mon-linux-x64.tgz -C build/linux_amd64 jvm-mon
