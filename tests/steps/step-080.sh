#!/usr/bin/env bash
# Test:
#   Cli tool: add/update README

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

accept_all_trust_prompts || exit 1

if ! "$GH_TEST_BIN/cli" installer; then
    echo "! Failed to execute the install script"
    exit 1
fi

mkdir -p "$GH_TEST_TMP/not/a/git/repo" && cd "$GH_TEST_TMP/not/a/git/repo" || exit 1

if "$GH_INSTALL_BIN_DIR/cli" readme add; then
    echo "! Expected to fail"
    exit 1
fi

mkdir -p "$GH_TEST_TMP/test080" &&
    cd "$GH_TEST_TMP/test080" &&
    git init || exit 1

"$GH_INSTALL_BIN_DIR/cli" readme update &&
    [ -f .githooks/README.md ] ||
    exit 1

if "$GH_INSTALL_BIN_DIR/cli" readme add; then
    echo "! Expected to fail"
    exit 1
fi
