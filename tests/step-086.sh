#!/bin/sh
# Test:
#   Cli tool: list Githooks configuration

TEST_DIR=$(cd "$(dirname "$0")" && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

acceptAllTrustPrompts || exit 1

if ! "$GH_TEST_BIN/cli" installer; then
    echo "! Failed to execute the install script"
    exit 1
fi

mkdir -p "$GH_TEST_TMP/test086" && cd "$GH_TEST_TMP/test086" || exit 3

! "$GH_INSTALL_BIN_DIR/cli" config list --local || exit 4 # not a Git repo

git init || exit 5

"$GH_INSTALL_BIN_DIR/cli" config update --enable || exit 7
"$GH_INSTALL_BIN_DIR/cli" config list
"$GH_INSTALL_BIN_DIR/cli" config list | grep -q -i 'githooks.autoUpdateEnabled' || exit 8
"$GH_INSTALL_BIN_DIR/cli" config list --global | grep -q -i 'githooks.autoUpdateEnabled' || exit 9
! "$GH_INSTALL_BIN_DIR/cli" config list --local | grep -q -i 'githooks.autoUpdateEnabled' || exit 10
