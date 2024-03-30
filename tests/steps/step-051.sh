#!/usr/bin/env bash
# Test:
#   Cli tool: print version number

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

accept_all_trust_prompts || exit 1

"$GH_TEST_BIN/cli" installer || exit 1

if ! "$GH_INSTALL_BIN_DIR/cli" --version | grep -qE ".*[0-9]+\.[0-9]+\.[0-9]+"; then
    echo "! Unexpected cli version output"
    exit 1
fi
