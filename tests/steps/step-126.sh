#!/usr/bin/env bash
# Test:
#   Run a simple install and verify a hook triggers properly

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

# run the default install
"$GH_TEST_BIN/githooks-cli" installer "${EXTRA_INSTALL_ARGS[@]}" || exit 1

mkdir -p "$GH_TEST_TMP/test2-clone" || exit 1

mkdir -p "$GH_TEST_TMP/test2" &&
    cd "$GH_TEST_TMP/test2" &&
    git init &&
    mkdir -p .githooks/post-index-change &&
    echo "echo 'From githooks' > '$GH_TEST_TMP/hooktest'" >.githooks/post-index-change/test ||
    exit 1

# Clone inside this repository
git clone "$GH_TEST_TMP/test2-clone" "$GH_TEST_TMP/test2-cloned"

# Hooks should not have been triggered simply because we are inside a clone.
if grep -q 'From githooks' "$GH_TEST_TMP/hooktest" 2>/dev/null; then
    echo "Expected hook to not run"
    exit 3
fi
