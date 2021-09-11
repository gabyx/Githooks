#!/bin/sh
# Test:
#   Cli tool: manage trust settings

TEST_DIR=$(cd "$(dirname "$0")" && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

acceptAllTrustPrompts || exit 1

if ! "$GH_TEST_BIN/cli" installer; then
    echo "! Failed to execute the install script"
    exit 1
fi

mkdir -p "$GH_TEST_TMP/test081" &&
    cd "$GH_TEST_TMP/test081" &&
    git init || exit 1

"$GH_INSTALL_BIN_DIR/cli" trust &&
    [ -f .githooks/trust-all ] &&
    [ "$(git config --local --get githooks.trustAll)" = "true" ] ||
    exit 1

"$GH_INSTALL_BIN_DIR/cli" trust revoke &&
    [ -f .githooks/trust-all ] &&
    [ "$(git config --local --get githooks.trustAll)" = "false" ] ||
    exit 2

"$GH_INSTALL_BIN_DIR/cli" trust delete &&
    [ ! -f .githooks/trust-all ] &&
    [ "$(git config --local --get githooks.trustAll)" = "false" ] ||
    exit 3

"$GH_INSTALL_BIN_DIR/cli" trust forget &&
    [ -z "$(git config --local --get githooks.trustAll)" ] &&
    "$GH_INSTALL_BIN_DIR/cli" trust forget ||
    exit 4

"$GH_INSTALL_BIN_DIR/cli" trust invalid && exit 5

exit 0
