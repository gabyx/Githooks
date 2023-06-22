#!/usr/bin/env bash
# Test:
#   Cli tool: manage trusted repository configuration

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

acceptAllTrustPrompts || exit 1

if ! "$GH_TEST_BIN/cli" installer; then
    echo "! Failed to execute the install script"
    exit 1
fi

mkdir -p "$GH_TEST_TMP/test093" && cd "$GH_TEST_TMP/test093" || exit 2

! "$GH_INSTALL_BIN_DIR/cli" config trust-all --accept || exit 3

git init || exit 4

! "$GH_INSTALL_BIN_DIR/cli" config trust-all || exit 5

"$GH_INSTALL_BIN_DIR/cli" config trust-all --accept &&
    "$GH_INSTALL_BIN_DIR/cli" config trust-all --accept | grep -q 'trusts all hooks' || exit 6

"$GH_INSTALL_BIN_DIR/cli" config trust-all --deny &&
    "$GH_INSTALL_BIN_DIR/cli" config trust-all --print | grep -q 'does not trust hooks' || exit 7

"$GH_INSTALL_BIN_DIR/cli" config trust-all --reset &&
    "$GH_INSTALL_BIN_DIR/cli" config trust-all --print | grep -q 'is not set' || exit 8
