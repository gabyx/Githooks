#!/usr/bin/env bash
# Test:
#   Test clone url and clone branch settings

TEST_DIR=$(cd "$(dirname "$0")" && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

acceptAllTrustPrompts || exit 1

if ! "$GH_TEST_BIN/cli" installer; then
    echo "! Failed to execute the install script"
    exit 1
fi

"$GH_INSTALL_BIN_DIR/cli" config clone-url --set "https://wuagadugu.git" || exit 1
"$GH_INSTALL_BIN_DIR/cli" config clone-url --print | grep -q "wuagadugu" || exit 2

if ! git config githooks.cloneUrl | grep -q "wuagadugu"; then
    echo "Expected clone url to be set" >&2
    exit 1
fi

"$GH_INSTALL_BIN_DIR/cli" config clone-branch --set "gaga" || exit 3
"$GH_INSTALL_BIN_DIR/cli" config clone-branch --print | grep -q "gaga" || exit 4

if ! git config githooks.cloneBranch | grep -q "gaga"; then
    echo "Expected clone branch to be set" >&2
    exit 1
fi

"$GH_INSTALL_BIN_DIR/cli" config clone-branch --reset || exit 5
