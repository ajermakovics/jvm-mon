#!/bin/bash

echo "Build and run from Go source"

echo > log
DIR=`pwd`
echo "Dir: $DIR"
#GOPATH=$DIR
#GOROOT=$DIR/src
go build -o build/ && echo Built && ./build/jvm-mon-go $* 

echo log:
tail -n 5 log
# go run *.go
