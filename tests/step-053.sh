#!/bin/sh
# Test:
#   Cli tool: list current hooks

TEST_DIR=$(cd "$(dirname "$0")" && pwd)
# shellcheck disable=SC1090
. "$TEST_DIR/general.sh"

acceptAllTrustPrompts || exit 1

"$GH_TEST_BIN/cli" installer || exit 1

mkdir -p "$GH_TEST_TMP/test053/.githooks/pre-commit" &&
    echo 'echo "Hello"' >"$GH_TEST_TMP/test053/.githooks/pre-commit/example" &&
    cd "$GH_TEST_TMP/test053" &&
    git init || exit 1

if ! "$GH_INSTALL_BIN_DIR/cli" list | grep "example" | grep "'untrusted'" | grep "'active'"; then
    echo "! Unexpected cli list output"
    exit 1
fi

git commit -m 'Test'

if ! "$GH_INSTALL_BIN_DIR/cli" list | grep "example" | grep "'trusted'" | grep "'active'"; then
    echo "! Unexpected cli list output"
    exit 1
fi
