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

regex="${1:-".*"}"
tags="${2:-"test_docker"}"

if [ -d /cover ]; then
    if ! go test -tags "$tags" -run "$regex" ./... -test.coverprofile /cover/tests.cov -covermode=count -coverpkg ./...; then
        echo "! Go testsuite reported errors." >&2
        exit 1
    fi
else
    if ! go test -tags "$tags" -run "$regex" -v ./...; then
        echo "! Go testsuite reported errors." >&2
        exit 1
    fi
fi

exit 0
