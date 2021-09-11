#!/usr/bin/env bash
# Test:
#   Direct runner execution: break on errors

TEST_DIR=$(cd "$(dirname "$0")" && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

acceptAllTrustPrompts || exit 1

mkdir -p "$GH_TEST_TMP/test13" &&
    cd "$GH_TEST_TMP/test13" &&
    git init || exit 1

mkdir -p .githooks &&
    echo 'exit 1' >.githooks/pre-commit &&
    "$GH_TEST_BIN/runner" "$(pwd)"/.git/hooks/pre-commit

if [ $? -ne 1 ]; then
    echo "! Expected the hooks to fail"
    exit 1
fi

rm .githooks/pre-commit &&
    mkdir .githooks/pre-commit &&
    echo 'exit 1' >.githooks/pre-commit/test &&
    "$GH_TEST_BIN/runner" "$(pwd)"/.git/hooks/pre-commit

if [ $? -ne 1 ]; then
    echo "! Expected the hooks to fail"
    exit 1
fi

echo 'exit 0' >.githooks/pre-commit/test &&
    "$GH_TEST_BIN/runner" "$(pwd)"/.git/hooks/pre-commit ||
    exit 1
