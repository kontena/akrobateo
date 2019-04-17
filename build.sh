#!/bin/sh

# To easily cross-compile binaries
go get github.com/mitchellh/gox

VERSION=${DRONE_TAG:-latest}
GIT_COMMIT=$(git rev-list -1 HEAD || echo 'dirrrty')

CGO_ENABLED=0 gox -output="output/akrobateo_{{.OS}}_{{.Arch}}" \
  -osarch="darwin/amd64 linux/amd64 linux/arm64" \
  -ldflags "-s -w  -X github.com/kontena/akrobateo/version.GitCommit=${GIT_COMMIT} -X github.com/kontena/akrobateo/version.Version=${VERSION}" \
  github.com/kontena/akrobateo/cmd/manager/
