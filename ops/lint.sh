#!/bin/bash

set -o errexit
set -o nounset

export CGO_ENABLED=0
export GO111MODULE=on

PACKAGES=('./twingate')

echo -n "Checking gofmt on ${PACKAGES[*]}: "
ERRS=$(find ${PACKAGES[*]} -type f -name \*.go | xargs gofmt -l 2>&1 || true)
if [ -n "${ERRS}" ]; then
    echo "FAIL - the following files need to be gofmt'ed:"
    for e in ${ERRS}; do
        echo "    $e"
    done
    echo
    exit 1
fi
echo "PASS"
echo

LINTER_CONFIG='./ops/golangci.yml'
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
