#!/usr/bin/env bash
# Test:
#   Execute a dry-run installation

TEST_DIR=$(cd "$(dirname "$0")" && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

acceptAllTrustPrompts || exit 1

mkdir -p "$GH_TEST_TMP/test10/a" &&
    cd "$GH_TEST_TMP/test10/a" &&
    git init || exit 1

echo "n
y
$GH_TEST_TMP
" | "$GH_TEST_BIN/cli" installer --stdin --dry-run || exit 1

mkdir -p "$GH_TEST_TMP/test10/b" &&
    cd "$GH_TEST_TMP/test10/b" &&
    git init || exit 1

if grep -q 'https://github.com/gabyx/githooks' "$GH_TEST_TMP/test10/a/.git/hooks/pre-commit"; then
    echo "! Hooks are unexpectedly installed in A"
    exit 1
fi

if grep -q 'https://github.com/gabyx/githooks' "$GH_TEST_TMP/test10/b/.git/hooks/pre-commit"; then
    echo "! Hooks are unexpectedly installed in B"
    exit 1
fi
