#!/usr/bin/env bash
# Test:
#   Direct runner execution: ignoring some hooks

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

accept_all_trust_prompts || exit 1

mkdir -p "$GH_TEST_TMP/test15" &&
    cd "$GH_TEST_TMP/test15" &&
    git init || exit 1

mkdir -p .githooks/pre-commit &&
    echo 'exit 1' >.githooks/pre-commit/test.first &&
    echo 'exit 1' >.githooks/pre-commit/test.second &&
    echo "echo 'Third was run' >> '$GH_TEST_TMP/test015.out'" >.githooks/pre-commit/test.third &&
    echo '#!/bin/sh' >.githooks/pre-commit/test.fourth &&
    echo "echo 'Fourth was run' >> '$GH_TEST_TMP/test015.out'" >>.githooks/pre-commit/test.fourth &&
    chmod +x .githooks/pre-commit/test.fourth &&
    echo 'patterns: - pre-commit/*first' >.githooks/.ignore.yaml &&
    echo 'patterns: - ./*second' >.githooks/pre-commit/.ignore.yaml &&
    "$GH_TEST_BIN/githooks-runner" "$(pwd)"/.git/hooks/pre-commit ||
    exit 1

grep -q 'Third was run' "$GH_TEST_TMP/test015.out" &&
    grep -q 'Fourth was run' "$GH_TEST_TMP/test015.out" ||
    exit 1
