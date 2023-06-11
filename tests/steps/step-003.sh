#!/usr/bin/env bash
# Test:
#   Run a simple install and verify multiple hooks trigger properly

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

acceptAllTrustPrompts || exit 1

# run the default install
"$GH_TEST_BIN/cli" installer || exit 1

mkdir -p "$GH_TEST_TMP/test3" &&
    cd "$GH_TEST_TMP/test3" &&
    git init || exit 1

# set up 2 pre-commit hooks, execute them and verify that they worked
mkdir -p .githooks/pre-commit &&
    echo "echo 'Hook-1' >> '$GH_TEST_TMP/multitest'" >.githooks/pre-commit/test1 &&
    echo "echo 'Hook-2' >> '$GH_TEST_TMP/multitest'" >.githooks/pre-commit/test2 ||
    exit 1

git commit -a -m 'Test' 2>/dev/null

grep -q 'Hook-1' "$GH_TEST_TMP/multitest" && grep -q 'Hook-2' "$GH_TEST_TMP/multitest"
