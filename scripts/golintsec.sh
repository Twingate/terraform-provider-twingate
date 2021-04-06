#!/bin/bash

LINTER_CONFIG='./golangci.yml'
PACKAGES=('.' './twingate')

if [ ! -f "${GOPATH}"/bin/golangci-lint ]; then
    # install last golangci-lint
    curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b $(go env GOPATH)/bin latest
fi

if [ ! -x "$(command -v gosec)" ]; then
    # install gosec v2.7.0
    curl -sfL https://raw.githubusercontent.com/securego/gosec/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.7.0
fi

for i in "${!PACKAGES[@]}"; do
  GO_PACKAGES[$i]=${PACKAGES[$i]}'/...'
done

echo "Running linters with config from ${LINTER_CONFIG} on ${GO_PACKAGES[*]}"
golangci-lint run -c ${LINTER_CONFIG} ${GO_PACKAGES[*]}

echo "Running gosec security checker on ${GO_PACKAGES[*]}"
gosec ${PACKAGES} ${GO_PACKAGES[*]}
