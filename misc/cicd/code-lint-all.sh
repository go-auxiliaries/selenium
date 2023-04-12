#!/bin/sh
set -e

CWD=$(dirname "$(readlink -f "$0")")
"$CWD"/code-lint.sh

echo "golangci-lint run ./..."
golangci-lint run ./...
