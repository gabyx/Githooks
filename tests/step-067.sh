#!/bin/sh
# Test:
#   Run an single-repo install in a directory that is not a Git repository

TEST_DIR=$(cd "$(dirname "$0")" && pwd)
# shellcheck disable=SC1090
. "$TEST_DIR/general.sh"

acceptAllTrustPrompts || exit 1

mkdir "$GH_TEST_TMP/not-a-git-repo" && cd "$GH_TEST_TMP/not-a-git-repo" || exit 1

if ! "$GH_TEST_BIN/cli" installer; then
    echo "! Expected to succeed"
    exit 1
fi

if "$GH_TEST_BIN/cli" install; then
    echo "! Install into current repo should have failed"
    exit 1
fi
