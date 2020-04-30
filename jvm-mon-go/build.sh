#!/bin/bash

./make-agent.sh

echo "Building"
go build -o build/

export GOOS=linux
export GOARCH=amd64

DIR="${GOOS}_${GOARCH}"
rm build/${DIR}/*
mkdir -p build/${DIR}

echo "Building $DIR"
go build -o build/${DIR}

rm build/*.tgz
tar cvzf build/jvm-mon-osx.tgz -C build jvm-mon-go
tar cvzf build/jvm-mon-linux64.tgz -C build/linux_amd64 jvm-mon-go

#Mac:
#GOOS=darwin
#GOARCH=amd64
