#!/usr/bin/env bash
# Test:
#   Execute a dry-run, non-interactive installation

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

accept_all_trust_prompts || exit 1

mkdir -p "$GH_TEST_TMP/test33/a" &&
    cd "$GH_TEST_TMP/test33/a" &&
    git init || exit 1

"$GH_TEST_BIN/githooks-cli" installer --dry-run --non-interactive || exit 1

mkdir -p "$GH_TEST_TMP/test33/b" &&
    cd "$GH_TEST_TMP/test33/b" &&
    git init || exit 1

check_no_local_install "$GH_TEST_TMP/test33/a"
check_no_local_install "$GH_TEST_TMP/test33/b"
