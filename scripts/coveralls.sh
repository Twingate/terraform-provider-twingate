#!/bin/sh

set -o errexit
set -o nounset

TEST_RESULTS=${TEST_RESULTS:-"./test_results"}

go run github.com/mattn/goveralls -coverprofile="${TEST_RESULTS}"/coverage.out -service=circleci -repotoken "${COVERALLS_TOKEN}"