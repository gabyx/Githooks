#!/bin/sh
# Test:
#   Run a simple install and verify a hook triggers properly

TEST_DIR=$(cd "$(dirname "$0")" && pwd)
# shellcheck disable=SC1090
. "$TEST_DIR/general.sh"

# run the default install
"$GH_TEST_BIN/cli" installer || exit 1

mkdir -p "$GH_TEST_TMP/test2" &&
    cd "$GH_TEST_TMP/test2" &&
    git init || exit 1

# add a pre-commit hook, execute and verify that it worked
mkdir -p .githooks/pre-commit &&
    echo "echo 'From githooks' > '$GH_TEST_TMP/hooktest'" >.githooks/pre-commit/test ||
    exit 1

git commit --allow-empty -m ''

if grep -q 'From githooks' "$GH_TEST_TMP/hooktest"; then
    echo "Expected hook to not run"
    exit 1
fi

acceptAllTrustPrompts || exit 1

git commit --allow-empty -m ''

if ! grep -q 'From githooks' "$GH_TEST_TMP/hooktest"; then
    echo "Expected hook to run"
    exit 1
fi
