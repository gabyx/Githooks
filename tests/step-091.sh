#!/bin/sh
# Test:
#   Cli tool: manage previous search directory configuration

TEST_DIR=$(cd "$(dirname "$0")" && pwd)
# shellcheck disable=SC1090
. "$TEST_DIR/general.sh"

acceptAllTrustPrompts || exit 1

if ! "$GH_TEST_BIN/cli" installer; then
    echo "! Failed to execute the install script"
    exit 1
fi

! "$GH_INSTALL_BIN_DIR/cli" config search-dir || exit 2
! "$GH_INSTALL_BIN_DIR/cli" config search-dir --set || exit 3
! "$GH_INSTALL_BIN_DIR/cli" config search-dir --set a b c || exit 4

"$GH_INSTALL_BIN_DIR/cli" config search-dir --set /prev/search/dir || exit 5
"$GH_INSTALL_BIN_DIR/cli" config search-dir --print | grep -q '/prev/search/dir' || exit 6

"$GH_INSTALL_BIN_DIR/cli" config search-dir --reset
"$GH_INSTALL_BIN_DIR/cli" config search-dir --print | grep -q -i 'previous search directory is not set' || exit 7
