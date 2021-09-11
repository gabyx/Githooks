#!/bin/sh
# Test:
#   Direct runner execution: fail on shared hooks

TEST_DIR=$(cd "$(dirname "$0")" && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

acceptAllTrustPrompts || exit 1

git config --global githooks.testingTreatFileProtocolAsRemote "true"

mkdir -p "$GH_TEST_TMP/shared/hooks-018.git/pre-commit" &&
    echo 'exit 1' >"$GH_TEST_TMP/shared/hooks-018.git/pre-commit/fail" &&
    cd "$GH_TEST_TMP/shared/hooks-018.git" &&
    git init &&
    git add . &&
    git commit -m 'Initial commit' ||
    exit 1

mkdir -p "$GH_TEST_TMP/test18" &&
    cd "$GH_TEST_TMP/test18" &&
    git init || exit 1

mkdir -p .githooks &&
    echo "urls: - file://$GH_TEST_TMP/shared/hooks-018.git" >.githooks/.shared.yaml &&
    "$GH_TEST_BIN/cli" shared update ||
    exit 1

"$GH_TEST_BIN/runner" "$(pwd)"/.git/hooks/pre-commit

if [ $? -ne 1 ]; then
    echo "! Expected to fail on shared hook execution"
    exit 1
fi
