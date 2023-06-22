#!/usr/bin/env bash
# Test:
#   Execute a dry-run, non-interactive installation

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

acceptAllTrustPrompts || exit 1

mkdir -p "$GH_TEST_TMP/test33/a" &&
    cd "$GH_TEST_TMP/test33/a" &&
    git init || exit 1

"$GH_TEST_BIN/cli" installer --dry-run --non-interactive || exit 1

mkdir -p "$GH_TEST_TMP/test33/b" &&
    cd "$GH_TEST_TMP/test33/b" &&
    git init || exit 1

if grep -q 'https://github.com/gabyx/githooks' "$GH_TEST_TMP/test33/a/.git/hooks/pre-commit"; then
    echo "! Hooks are unexpectedly installed in A"
    exit 1
fi

if grep -q 'https://github.com/gabyx/githooks' "$GH_TEST_TMP/test33/b/.git/hooks/pre-commit"; then
    echo "! Hooks are unexpectedly installed in B"
    exit 1
fi
