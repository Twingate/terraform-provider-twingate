#!/bin/sh

set -o nounset

PACKAGE_NAME=./twingate/...
TEST_RESULTS=${TEST_RESULTS:-"./test_results"}

mkdir -p "${TEST_RESULTS}"

echo PACKAGE_NAME: "$PACKAGE_NAME"
echo "Running tests:"
go tool gotest.tools/gotestsum --packages "${PACKAGE_NAME}" --jsonfile "${TEST_RESULTS}"/report.json -- -coverpkg="${PACKAGE_NAME}" -coverprofile="${TEST_RESULTS}"/coverage.out "${PACKAGE_NAME}"

if [ $? -ne 0 ]; then
  set -o errexit

  echo "Retry failed tests:"
  grep '"Action":"fail"' "${TEST_RESULTS}"/report.json | grep -o '"Test":"[^"]*"' | awk -F':' '{print $2}' | tr -d '"' > "${TEST_RESULTS}"/failed_tests.txt
  go tool gotest.tools/gotestsum --rerun-fails=5 --packages "${PACKAGE_NAME}" -- -run "$(paste -sd "|" "${TEST_RESULTS}"/failed_tests.txt)" -coverpkg="${PACKAGE_NAME}" -coverprofile="${TEST_RESULTS}"/retry_coverage.out "${PACKAGE_NAME}"

  go tool github.com/wadey/gocovmerge "${TEST_RESULTS}"/coverage.out "${TEST_RESULTS}"/retry_coverage.out > "${TEST_RESULTS}"/final_coverage.out

else

  go tool github.com/wadey/gocovmerge "${TEST_RESULTS}"/coverage.out "${TEST_RESULTS}"/coverage.out > "${TEST_RESULTS}"/final_coverage.out || exit 1

fi
