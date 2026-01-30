#!/bin/bash
cd "$(dirname "$0")"
set -e
cd ..
export GOBIN=$GOPATH/bin
export PATH=$GOBIN:$PATH
[[ ! -d "${GOBIN}" ]] && mkdir -p "${GOBIN}"
export GOBIN=$PWD/bin
export PATH="${GOBIN}:${PATH}"
echo "GOBIN=${GOBIN}"
rm -rf tmp
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b "$GOBIN" v1.64.7
