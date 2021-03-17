#!/bin/sh
# Test:
#   Direct runner execution: update a hook in a trusted repository

TEST_DIR=$(cd "$(dirname "$0")" && pwd)
# shellcheck disable=SC1090
. "$TEST_DIR/general.sh"

mkdir -p "$GH_TEST_TMP/test34" &&
    cd "$GH_TEST_TMP/test34" &&
    git init || exit 1

mkdir -p .githooks/pre-commit &&
    touch .githooks/trust-all &&
    echo "echo 'Trusted hook' > '$GH_TEST_TMP/test34.out'" >.githooks/pre-commit/test &&
    TRUST_ALL_HOOKS=Y ACCEPT_CHANGES=N \
        "$GH_TEST_BIN/runner" "$(pwd)"/.git/hooks/pre-commit

if ! grep -q "Trusted hook" "$GH_TEST_TMP/test34.out"; then
    echo "! Expected hook was not run"
    exit 1
fi

echo "echo 'Changed hook' > '$GH_TEST_TMP/test34.out'" >.githooks/pre-commit/test &&
    TRUST_ALL_HOOKS="" ACCEPT_CHANGES=N \
        "$GH_TEST_BIN/runner" "$(pwd)"/.git/hooks/pre-commit

if ! grep -q "Changed hook" "$GH_TEST_TMP/test34.out"; then
    echo "! Changed hook was not run"
    exit 1
fi
