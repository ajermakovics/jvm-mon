#!/bin/bash

echo "Build and run from Go source"

echo > log
DIR=`pwd`
echo "Dir: $DIR"

go build -o build/ && echo Built && ./build/jvm-mon-go $*

echo log:
tail -n 5 log
# go run *.go
