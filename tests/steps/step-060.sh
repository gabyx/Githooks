#!/usr/bin/env bash
# Test:
#   Cli tool: list shows files in trusted repos

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

acceptAllTrustPrompts || exit 1

"$GH_TEST_BIN/cli" installer || exit 1

mkdir -p "$GH_TEST_TMP/test060/.githooks/pre-commit" &&
    echo 'echo "Hello"' >"$GH_TEST_TMP/test060/.githooks/pre-commit/first" &&
    echo 'echo "Hello"' >"$GH_TEST_TMP/test060/.githooks/pre-commit/second" &&
    touch "$GH_TEST_TMP/test060/.githooks/trust-all" &&
    cd "$GH_TEST_TMP/test060" &&
    git init &&
    git config --local githooks.trustAll true ||
    exit 1

if ! "$GH_INSTALL_BIN_DIR/cli" list | grep "first" | grep -q "'trusted'"; then
    echo "! Unexpected cli list output (1)"
    exit 1
fi

if ! "$GH_INSTALL_BIN_DIR/cli" list | grep "second" | grep -q "'trusted'"; then
    echo "! Unexpected cli list output (2)"
    exit 1
fi
