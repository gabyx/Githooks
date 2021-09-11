#!/usr/bin/env bash
# Test:
#   Cli tool: list current hooks per type

TEST_DIR=$(cd "$(dirname "$0")" && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

acceptAllTrustPrompts || exit 1

"$GH_TEST_BIN/cli" installer || exit 1

mkdir -p "$GH_TEST_TMP/test054/.githooks/pre-commit" &&
    mkdir -p "$GH_TEST_TMP/test054/.githooks/post-commit" &&
    echo 'echo "Hello"' >"$GH_TEST_TMP/test054/.githooks/pre-commit/pre-example" &&
    echo 'echo "Hello"' >"$GH_TEST_TMP/test054/.githooks/post-commit/post-example" &&
    cd "$GH_TEST_TMP/test054" &&
    git init || exit 1

if ! git hooks list pre-commit | grep "pre-example"; then
    echo "! Unexpected cli list output"
    exit 1
fi

if ! git hooks list post-commit | grep "post-example"; then
    echo "! Unexpected cli list output"
    exit 1
fi

if ! git hooks list post-commit | grep -v "pre-example"; then
    echo "! Unexpected cli list output"
    exit 1
fi
