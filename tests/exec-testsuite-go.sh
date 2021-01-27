#!/bin/sh

if ! grep '/docker/' </proc/self/cgroup >/dev/null 2>&1; then
    echo "! This script is only meant to be run in a Docker container"
    exit 1
fi

DIR="$(cd "$(dirname "$0")" >/dev/null 2>&1 && pwd)"

REPO_DIR="$DIR/.."
GO_SRC="$REPO_DIR/githooks"

cd "$GO_SRC" || exit 1

echo "Go generate ..."

export CGO_ENABLED=0

go mod vendor
go generate -mod vendor ./...

if ! go test ./...; then
    echo "Go testsuite reported errors."
    exit 1
fi

exit 0
