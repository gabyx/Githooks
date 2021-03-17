#!/bin/sh
# Test:
#   Direct runner execution: test a single pre-commit hook file

TEST_DIR=$(cd "$(dirname "$0")" && pwd)
# shellcheck disable=SC1090
. "$TEST_DIR/general.sh"

acceptAllTrustPrompts || exit 1

mkdir -p "$GH_TEST_TMP/test12" &&
    cd "$GH_TEST_TMP/test12" &&
    git init || exit 1

mkdir -p .githooks &&
    echo "echo 'Direct execution' > '$GH_TEST_TMP/test012.out'" >.githooks/pre-commit &&
    "$GH_TEST_BIN/runner" "$(pwd)"/.git/hooks/pre-commit ||
    exit 1

grep -q 'Direct execution' "$GH_TEST_TMP/test012.out"
