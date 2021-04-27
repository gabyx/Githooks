#!/bin/sh
# Test:
#   Cli tool: manage update time configuration

TEST_DIR=$(cd "$(dirname "$0")" && pwd)
# shellcheck disable=SC1090
. "$TEST_DIR/general.sh"

acceptAllTrustPrompts || exit 1

if ! "$GH_TEST_BIN/cli" installer; then
    echo "! Failed to execute the install script"
    exit 1
fi

! "$GH_INSTALL_BIN_DIR/cli" config update-time || exit 2

"$GH_INSTALL_BIN_DIR/cli" config update-time --print | grep -q 'never' || exit 3

git config --global githooks.autoUpdateCheckTimestamp 123 &&
    "$GH_INSTALL_BIN_DIR/cli" config update-time --print | grep -q 'never' && exit 4

"$GH_INSTALL_BIN_DIR/cli" config update-time --reset &&
    "$GH_INSTALL_BIN_DIR/cli" config update-time --print | grep -q 'never' || exit 5
