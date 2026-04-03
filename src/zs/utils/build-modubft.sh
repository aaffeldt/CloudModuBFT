#!/bin/bash
main_file=$1
main_file_folder=$(dirname "$main_file")
main_file_basename=$(basename "$main_file")
binary="/tmp/cli"

pushd "$main_file_folder" || exit 1

CGO_ENABLED=0  GOOS=linux GOARCH=amd64 go version

echo "Building $main_file_basename to $binary"
CGO_ENABLED=0  GOOS=linux GOARCH=amd64 go build  -o "$binary"  "$main_file_basename" || exit 1

popd  || exit 1

echo "$binary"
