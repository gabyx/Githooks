#!/usr/bin/env bash
# Test:
#   Cli tool: manage update state configuration

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

accept_all_trust_prompts || exit 1

if ! "$GH_TEST_BIN/cli" installer; then
    echo "! Failed to execute the install script"
    exit 1
fi

! "$GH_INSTALL_BIN_DIR/cli" config update || exit 2

"$GH_INSTALL_BIN_DIR/cli" config update --disable &&
    "$GH_INSTALL_BIN_DIR/cli" config update --print | grep -q 'disabled' || exit 3

"$GH_INSTALL_BIN_DIR/cli" config update --enable &&
    "$GH_INSTALL_BIN_DIR/cli" config update --print | grep -q 'enabled' || exit 4

"$GH_INSTALL_BIN_DIR/cli" config update --disable &&
    "$GH_INSTALL_BIN_DIR/cli" config update --print | grep -q 'disabled' || exit 5
