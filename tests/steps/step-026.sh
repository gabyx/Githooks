#!/usr/bin/env bash
# Test:
#   Direct runner execution: do not accept any new hooks

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

mkdir -p "$GH_TEST_TMP/test26" &&
    cd "$GH_TEST_TMP/test26" &&
    git init || exit 1

mkdir -p .githooks &&
    mkdir -p .githooks/pre-commit &&
    echo "echo 'First execution' >> '$GH_TEST_TMP/test026.out'" >.githooks/pre-commit/test &&
    ACCEPT_CHANGES=N "$GH_TEST_BIN/runner" "$(pwd)"/.git/hooks/pre-commit

if grep -q "First execution" "$GH_TEST_TMP/test026.out"; then
    echo "! Expected to refuse executing the hook"
    exit 1
fi
