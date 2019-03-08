#!/bin/sh

# To easily cross-compile binaries
go get github.com/mitchellh/gox

VERSION=${DRONE_TAG:-head}
GIT_COMMIT=$(git rev-list -1 HEAD || echo 'dirrrty')

CGO_ENABLED=0 gox -output="output/service-lb-operator_{{.OS}}_{{.Arch}}" \
  -osarch="darwin/amd64 linux/amd64 linux/arm64" \
  -ldflags "-s -w  -X github.com/kontena/service-lb-operator/version.GitCommit=${GIT_COMMIT} -X github.com/kontena/service-lb-operator/version.Version=${VERSION}" \
  github.com/kontena/service-lb-operator/cmd/manager/
