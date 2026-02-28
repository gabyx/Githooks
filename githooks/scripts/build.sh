#!/usr/bin/env bash
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
LD_FLAGS=()

export CGO_ENABLED=0

parse_args() {
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
            LD_FLAGS+=("-ldflags" "-s -w") # strip debug information
        elif [ "$p" = "--help" ] || [ "$p" = "-h" ]; then
            echo "Usage 'build.sh'
                [--bin-dir <dir>]
                [--build-tags <tags>]
                [--coverage]
                [--prod]"
            exit 0
        else
            echo "! Unknown argument \`$p\`" >&2
            return 1
        fi
        prev_p="$p"
    done
}

parse_args "$@"

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

# Add debug tag if set
if [ -n "$DEBUG_TAG" ] &&
    ! echo "$BUILD_TAGS" | grep -q "$DEBUG_TAG"; then
    BUILD_TAGS="$BUILD_TAGS,$DEBUG_TAG"
fi

if [ -z "$BUILD_COVERAGE" ]; then
    echo "Build normal ..."
    echo "go install ..."

    cmd=(go generate -mod=vendor ./...)
    echo -e "Generating with:\n" "${cmd[@]}"
    "${cmd[@]}"

    # shellcheck disable=SC2086
    cmd=(go install -mod=vendor -tags "$BUILD_TAGS" "${LD_FLAGS[@]}" ./...)
    echo -e "Building with:\n" "${cmd[@]}"
    "${cmd[@]}"

    mv "$GOBIN/cli" "$GOBIN/githooks-cli"
    mv "$GOBIN/runner" "$GOBIN/githooks-runner"
    mv "$GOBIN/dialog" "$GOBIN/githooks-dialog"
else
    echo "Build coverage ..."
    BUILD_TAGS="$BUILD_TAGS,coverage"

    echo "go test ..."
    go generate -mod=vendor ./...
    # shellcheck disable=SC2086
    go test ./apps/cli -tags "$BUILD_TAGS" -covermode=count -coverpkg ./... -c -o "$GOBIN/githooks-cli"
    # shellcheck disable=SC2086
    go test ./apps/dialog -tags "$BUILD_TAGS" -covermode=count -coverpkg ./... -c -o "$GOBIN/githooks-dialog"
    # shellcheck disable=SC2086
    go test ./apps/runner -tags "$BUILD_TAGS" -covermode=count -coverpkg ./... -c -o "$GOBIN/githooks-runner"
fi
