#!/bin/sh
# Test:
#   Direct runner execution: test a shared repo with checked in compiled hooks

TEST_DIR=$(cd "$(dirname "$0")" && pwd)
# shellcheck disable=SC1090
. "$TEST_DIR/general.sh"

"$GH_TEST_BIN/cli" installer || exit 1
acceptAllTrustPrompts || exit 1

cleanup() {
    true
}

trap cleanup EXIT

# Make our pre-compiled shared hook repo.
mkdir -p "$GH_TEST_TMP/shared" &&
    cd "$GH_TEST_TMP/shared" &&
    git init || exit 3

# Make folders.
mkdir -p "githooks/pre-commit" "dist" || exit 4

# Git LFS (if available)
if ! git-lfs --version; then
    git lfs track "*.exe"
fi

# Make runner script.
mkdir -p .githooks &&
    cat <<"EOF" >"githooks/pre-commit/custom.yaml" || exit 5
cmd: "dist/custom-${env:GITHOOKS_OS}-${env:GITHOOKS_ARCH}.exe"
version: 1
EOF

# Make the hook source file.
cat <<"EOF" >"custom.go" || exit 5
package main

import (
    "fmt"
    "runtime"
)

func main() {
    fmt.Printf("%s\n%s\n%s", runtime.GOOS, runtime.GOARCH, "Hello from compiled hook")
}
EOF

# Detect the os/arch.
OUT=$(go run custom.go) || exit 6
OS=$(echo "$OUT" | head -1 | tail -1) || exit 7
ARCH=$(echo "$OUT" | head -2 | tail -1) || exit 8

env GOOS="$OS" GOARCH="$ARCH" \
    go build -o "dist/custom-$OS-$ARCH.exe" custom.go || exit 9
git add . &&
    git commit -a -m "built hooks" || exit 10

# Make normal repo.
mkdir -p "$GH_TEST_TMP/test123" &&
    cd "$GH_TEST_TMP/test123" &&
    git init || exit 11

# Add the shared repo
"$GITHOOKS_INSTALL_BIN_DIR/cli" shared add --local "file://$GH_TEST_TMP/shared" || exit 12
"$GITHOOKS_INSTALL_BIN_DIR/cli" shared update || exit 13

# Execute pre-commit by the runner
OUT=$("$GH_TEST_BIN/runner" "$(pwd)"/.git/hooks/pre-commit 2>&1)
# shellcheck disable=SC2181,SC2016
if [ "$?" -ne 0 ] ||
    ! echo "$OUT" | grep "Hello from compiled hook"; then
    echo "! Expected compiled to be executed."
    echo "$OUT"
    exit 14
fi
