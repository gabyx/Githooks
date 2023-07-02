#!/usr/bin/env bash
set -e
set -u

if [ "$DOCKER_RUNNING" != "true" ]; then
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

if [ -d /cover ]; then
    if ! go test ./... -test.coverprofile /cover/tests.cov -covermode=count -coverpkg ./...; then
        echo "! Go testsuite reported errors." >&2
        exit 1
    fi
else
    if ! go test -v ./...; then
        echo "! Go testsuite reported errors." >&2
        exit 1
    fi
fi

exit 0
