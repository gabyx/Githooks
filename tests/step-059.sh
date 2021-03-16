#!/bin/sh
# Test:
#   Cli tool: list shows ignored files

TEST_DIR=$(cd "$(dirname "$0")" && pwd)
# shellcheck disable=SC1090
. "$TEST_DIR/general.sh"

acceptAllTrustPrompts || exit 1

"$GH_TEST_BIN/cli" installer || exit 1

mkdir -p "$GH_TEST_TMP/test059/.githooks/pre-commit" &&
    echo 'echo "Hello"' >"$GH_TEST_TMP/test059/.githooks/pre-commit/first" &&
    echo 'echo "Hello"' >"$GH_TEST_TMP/test059/.githooks/pre-commit/second" &&
    echo 'patterns: - pre-commit/first' >"$GH_TEST_TMP/test059/.githooks/.ignore.yaml" &&
    echo 'patterns: - pre-commit/second' >"$GH_TEST_TMP/test059/.githooks/pre-commit/.ignore.yaml" &&
    cd "$GH_TEST_TMP/test059" &&
    git init || exit 1

if ! "$GITHOOKS_INSTALL_BIN_DIR/cli" list | grep "first" | grep -q "'ignored'"; then
    echo "! Unexpected cli list output (1)"
    exit 1
fi

if ! "$GITHOOKS_INSTALL_BIN_DIR/cli" list | grep "second" | grep -q "'ignored'"; then
    echo "! Unexpected cli list output (2)"
    exit 1
fi
