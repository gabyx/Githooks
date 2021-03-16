#!/bin/sh
# Test:
#   Cli tool: manage trusted repository configuration

TEST_DIR=$(cd "$(dirname "$0")" && pwd)
# shellcheck disable=SC1090
. "$TEST_DIR/general.sh"

acceptAllTrustPrompts || exit 1

if ! "$GH_TEST_BIN/cli" installer; then
    echo "! Failed to execute the install script"
    exit 1
fi

mkdir -p "$GH_TEST_TMP/test093" && cd "$GH_TEST_TMP/test093" || exit 2

! "$GITHOOKS_INSTALL_BIN_DIR/cli" config trust-all --accept || exit 3

git init || exit 4

! "$GITHOOKS_INSTALL_BIN_DIR/cli" config trust-all || exit 5

"$GITHOOKS_INSTALL_BIN_DIR/cli" config trust-all --accept &&
    "$GITHOOKS_INSTALL_BIN_DIR/cli" config trust-all --accept | grep -q 'trusts all hooks' || exit 6

"$GITHOOKS_INSTALL_BIN_DIR/cli" config trust-all --deny &&
    "$GITHOOKS_INSTALL_BIN_DIR/cli" config trust-all --print | grep -q 'does not trust hooks' || exit 7

"$GITHOOKS_INSTALL_BIN_DIR/cli" config trust-all --reset &&
    "$GITHOOKS_INSTALL_BIN_DIR/cli" config trust-all --print | grep -q 'is not set' || exit 8
