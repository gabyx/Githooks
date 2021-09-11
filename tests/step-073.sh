#!/usr/bin/env bash
# Test:
#   Run the cli tool trying to list a not yet trusted repo

if ! "$GH_TEST_BIN/cli" installer; then
    echo "! Failed to execute the install script"
    exit 1
fi

TEST_DIR=$(cd "$(dirname "$0")" && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

acceptAllTrustPrompts || exit 1

mkdir -p "$GH_TEST_TMP/test073/.githooks/pre-commit" &&
    echo 'echo "Hello"' >"$GH_TEST_TMP/test073/.githooks/pre-commit/testing" &&
    touch "$GH_TEST_TMP/test073/.githooks/trust-all" &&
    cd "$GH_TEST_TMP/test073" &&
    git init || exit 1

if "$GH_INSTALL_BIN_DIR/cli" list pre-commit | grep -i "'trusted'"; then
    echo "! Unexpected list result"
    exit 1
fi
