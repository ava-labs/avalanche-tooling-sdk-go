#!/usr/bin/env bash
# Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
# See the file LICENSE for licensing terms.

set -o errexit
set -o nounset
set -o pipefail

SDK_PATH=$(
    cd "$(dirname "${BASH_SOURCE[0]}")"
    cd .. && pwd
)

GOLANGCI_LINT_VERSION=v2.11.4

# avoid calling go install unless it is needed: makes the script able to be used offline

exists=true
which golangci-lint > /dev/null 2>&1 || exists=false

install=false
if [ $exists = true ]
then
	golangci-lint --version | grep $GOLANGCI_LINT_VERSION > /dev/null 2>&1 || install=true
else
	install=true
fi

if [ $install = true ]
then
	go install -v github.com/golangci/golangci-lint/v2/cmd/golangci-lint@${GOLANGCI_LINT_VERSION}
fi

echo "Running general lint..."
golangci-lint run ./... --timeout 5m

echo ""
echo "Running lint for example files..."

# Lint each example file separately (they each have their own main function)
find "$SDK_PATH/examples" -name "*.go" -type f | while read -r file; do
    echo "  Linting $file..."
    golangci-lint run "$file" --timeout 1m
done

echo ""
echo "Lint completed successfully!"
