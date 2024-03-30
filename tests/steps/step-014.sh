#!/usr/bin/env bash
# Test:
#   Direct runner execution: disable running custom hooks

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

accept_all_trust_prompts || exit 1

mkdir -p "$GH_TEST_TMP/test14" &&
    cd "$GH_TEST_TMP/test14" &&
    git init || exit 1

mkdir -p .githooks/pre-commit &&
    echo 'exit 1' >.githooks/pre-commit/test &&
    "$GH_TEST_BIN/runner" "$(pwd)"/.git/hooks/pre-commit

if [ $? -ne 1 ]; then
    echo "! Expected the hooks to fail"
    exit 1
fi

GITHOOKS_DISABLE=1 "$GH_TEST_BIN/runner" "$(pwd)"/.git/hooks/pre-commit ||
    exit 1
