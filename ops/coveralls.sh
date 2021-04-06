#!/bin/sh
export GO111MODULE=on

TEST_RESULTS=${TEST_RESULTS:-"./test/out"}

GO111MODULE=off go get golang.org/x/tools/cmd/cover
GO111MODULE=off go get github.com/mattn/goveralls

goveralls -coverprofile="${TEST_RESULTS}"/coverage.out -service=circleci -repotoken "${COVERALLS_TOKEN}"