#!/bin/sh
DIR="$(cd "$(dirname "$0")" >/dev/null 2>&1 && pwd)"
GO_SRC="$DIR/.."

set -e
set -u

die() {
    echo "!! " "$@" >&2
    exit 1
}

BIN_DIR=""
BUILD_TAGS=""
BUILD_COVERAGE=""
DEBUG_TAG="debug"

export CGO_ENABLED=0

parseArgs() {
    prev_p=""
    for p in "$@"; do
        if [ "$p" = "--bin-dir" ]; then
            true
        elif [ "$prev_p" = "--bin-dir" ]; then
            BIN_DIR="$p"
        elif [ "$p" = "--build-tags" ]; then
            true
        elif [ "$prev_p" = "--build-tags" ]; then
            BUILD_TAGS="$p"
        elif [ "$p" = "--coverage" ]; then
            BUILD_COVERAGE="true"
        elif [ "$p" = "--prod" ]; then
            DEBUG_TAG=""
        else
            echo "! Unknown argument \`$p\`" >&2
            return 1
        fi
        prev_p="$p"
    done
}

parseArgs "$@" || die "Parsing args failed."

cd "$GO_SRC"

export GOBIN="$GO_SRC/bin"
if [ -n "$BIN_DIR" ]; then
    if [ -d "$BIN_DIR" ]; then
        rm -rf "$BIN_DIR" || true
    fi
    export GOBIN="$BIN_DIR"
fi

if [ ! -d "$GO_SRC/vendor" ]; then
    echo "go vendor ..."
    go mod vendor
fi

# Add debug tag it set
if [ -n "$BUILD_TAGS" ]; then
    BUILD_TAGS="$BUILD_TAGS,$DEBUG_TAG"
fi

if [ -z "$BUILD_COVERAGE" ]; then
    echo "Build normal ..."
    echo "go install ..."
    go generate -mod=vendor ./...
    # shellcheck disable=SC2086
    go install -mod=vendor -tags "$BUILD_TAGS" ./...
else
    echo "Build coverage ..."
    BUILD_TAGS="$BUILD_TAGS,coverage"

    echo "go test ..."
    go generate -mod=vendor ./...
    # shellcheck disable=SC2086
    go test ./apps/cli -tags "$BUILD_TAGS" -covermode=count -coverpkg ./... -c -o "$GOBIN/cli"
    # shellcheck disable=SC2086
    go test ./apps/dialog -tags "$BUILD_TAGS" -covermode=count -coverpkg ./... -c -o "$GOBIN/dialog"
    # shellcheck disable=SC2086
    go test ./apps/runner -tags "$BUILD_TAGS" -covermode=count -coverpkg ./... -c -o "$GOBIN/runner"
fi
