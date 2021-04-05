#!/usr/bin/bash

gofmt_files=$(gofmt -l `find . -name '*.go' | grep -v vendor`)
if [[ -n ${gofmt_files} ]]; then
    echo 'Check the following files:'
    echo "${gofmt_files}"
    echo "You can use the command: \`make fmt\` locally."
    exit 1
fi

exit 0