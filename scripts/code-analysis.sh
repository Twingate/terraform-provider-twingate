#!/bin/bash

set -o errexit
set -o nounset

export CGO_ENABLED=0
export GO111MODULE=on

PACKAGES=('./twingate')

LINTER_CONFIG='./scripts/golangci.yml'

if [ ! -f "${GOPATH}"/bin/golangci-lint ]; then
    # install last golangci-lint
    curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b $(go env GOPATH)/bin latest
fi

for i in "${!PACKAGES[@]}"; do
  GO_PACKAGES[$i]=${PACKAGES[$i]}'/...'
done

echo "Running linters with config from ${LINTER_CONFIG} on ${GO_PACKAGES[*]}"
golangci-lint run -c ${LINTER_CONFIG} ${GO_PACKAGES[*]}

