#!/bin/sh

set -o nounset

PACKAGE_NAME=./twingate/...
TEST_RESULTS=${TEST_RESULTS:-"./test_results"}

mkdir -p "${TEST_RESULTS}"

MAIN_COVDIR="${TEST_RESULTS}/covdata-main"
RETRY_COVDIR="${TEST_RESULTS}/covdata-retry"
MERGED_COVDIR="${TEST_RESULTS}/covdata-merged"
rm -rf "${MAIN_COVDIR}" "${RETRY_COVDIR}" "${MERGED_COVDIR}"
mkdir -p "${MAIN_COVDIR}" "${RETRY_COVDIR}" "${MERGED_COVDIR}"

echo PACKAGE_NAME: "$PACKAGE_NAME"
echo "Running tests:"
go tool gotest.tools/gotestsum --packages "${PACKAGE_NAME}" --jsonfile "${TEST_RESULTS}"/report.json \
  -- -cover -coverpkg="${PACKAGE_NAME}" "${PACKAGE_NAME}" -args -test.gocoverdir="${MAIN_COVDIR}"

if [ $? -ne 0 ]; then
  set -o errexit

  echo "Retry failed tests:"
  grep '"Action":"fail"' "${TEST_RESULTS}"/report.json | grep -o '"Test":"[^"]*"' | awk -F':' '{print $2}' | tr -d '"' > "${TEST_RESULTS}"/failed_tests.txt
  go tool gotest.tools/gotestsum --rerun-fails=5 --packages "${PACKAGE_NAME}" \
    -- -run "$(paste -sd "|" "${TEST_RESULTS}"/failed_tests.txt)" -cover -coverpkg="${PACKAGE_NAME}" "${PACKAGE_NAME}" \
       -args -test.gocoverdir="${RETRY_COVDIR}"

  go tool covdata merge -i="${MAIN_COVDIR},${RETRY_COVDIR}" -o="${MERGED_COVDIR}"
  go tool covdata textfmt -i="${MERGED_COVDIR}" -o="${TEST_RESULTS}"/final_coverage.out
else

  go tool covdata textfmt -i="${MAIN_COVDIR}" -o="${TEST_RESULTS}"/final_coverage.out

fi
