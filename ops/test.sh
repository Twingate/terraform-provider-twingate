#!/bin/sh

set -o errexit
set -o nounset

export CGO_ENABLED=1
export GO111MODULE=on

PACKAGE_NAMES=$(go list ./twingate/...)
TEST_RESULTS=${TEST_RESULTS:-"./test/out"}

mkdir -p "${TEST_RESULTS}"

echo PACKAGE_NAMES: "$PACKAGE_NAMES"
echo "Running tests:"
gotestsum --format standard-quiet --junitfile "${TEST_RESULTS}"/test-results.xml -- -coverpkg=./... -coverprofile="${TEST_RESULTS}"/coverage.out.tmp ${PACKAGE_NAMES}
echo

echo "Generating coverage report (removing generated **/api/gen/** and *.pb.go files)"
grep -f ./ops/coverage_ignore_patterns -v "${TEST_RESULTS}"/coverage.out.tmp > "${TEST_RESULTS}"/coverage.out
go tool cover -html="${TEST_RESULTS}"/coverage.out -o "${TEST_RESULTS}"/coverage.html
echo
