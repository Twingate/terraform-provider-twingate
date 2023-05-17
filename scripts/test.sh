#!/bin/sh

set -o errexit
set -o nounset

PACKAGE_NAME=./twingate/...
TEST_RESULTS=${TEST_RESULTS:-"./test_results"}

mkdir -p "${TEST_RESULTS}"

echo PACKAGE_NAME: "$PACKAGE_NAME"
echo "Running tests:"
# todo: change --format standard-quiet
go run gotest.tools/gotestsum --rerun-fails --packages "${PACKAGE_NAME}" --format testname --junitfile "${TEST_RESULTS}"/test-results.xml -- -coverpkg="${PACKAGE_NAME}" -coverprofile="${TEST_RESULTS}"/coverage.out.tmp "${PACKAGE_NAME}"
echo

echo "Generating coverage report (removing generated **/api/gen/** and *.pb.go files)"
grep -f ./scripts/coverage_ignore_patterns -v "${TEST_RESULTS}"/coverage.out.tmp > "${TEST_RESULTS}"/coverage.out
go tool cover -html="${TEST_RESULTS}"/coverage.out -o "${TEST_RESULTS}"/coverage.html
echo
