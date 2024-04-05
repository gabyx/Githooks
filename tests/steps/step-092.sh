#!/usr/bin/env bash
# Test:
#   Cli tool: manage global shared hook repository configuration

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

accept_all_trust_prompts || exit 1

if ! "$GH_TEST_BIN/githooks-cli" installer; then
    echo "! Failed to execute the install script"
    exit 1
fi

! "$GH_INSTALL_BIN_DIR/githooks-cli" config shared || exit 1
! "$GH_INSTALL_BIN_DIR/githooks-cli" config shared --add || exit 1
! "$GH_INSTALL_BIN_DIR/githooks-cli" config shared --local --add "asd" || exit 1

mkdir -p "$GH_TEST_TMP/test092" &&
    cd "$GH_TEST_TMP/test092" &&
    git init || exit 2

! "$GH_INSTALL_BIN_DIR/githooks-cli" config shared --local --add "" || exit 3
! "$GH_INSTALL_BIN_DIR/githooks-cli" config shared --global --local --add "a" "b" || exit 3
! "$GH_INSTALL_BIN_DIR/githooks-cli" config shared --local --print --add "a" "b" || exit 3

"$GH_INSTALL_BIN_DIR/githooks-cli" config shared --add "file://$GH_TEST_TMP/test/repo1.git" "file://$GH_TEST_TMP/test/repo2.git" || exit 4
"$GH_INSTALL_BIN_DIR/githooks-cli" config shared --global --print | grep -q 'test/repo1' || exit 5
"$GH_INSTALL_BIN_DIR/githooks-cli" config shared --global --print | grep -q 'test/repo2' || exit 6
! "$GH_INSTALL_BIN_DIR/githooks-cli" config shared --local --print | grep -q 'test/repo' || exit 7

"$GH_INSTALL_BIN_DIR/githooks-cli" config shared --local --add "file://$GH_TEST_TMP/test/repo3.git" || exit 8
! "$GH_INSTALL_BIN_DIR/githooks-cli" config shared --global --print | grep -q 'test/repo3' || exit 9
"$GH_INSTALL_BIN_DIR/githooks-cli" config shared --local --print | grep -q 'test/repo3' || exit 10
"$GH_INSTALL_BIN_DIR/githooks-cli" config shared --print | grep -q 'test/repo1' || exit 11
"$GH_INSTALL_BIN_DIR/githooks-cli" config shared --print | grep -q 'test/repo2' || exit 12
"$GH_INSTALL_BIN_DIR/githooks-cli" config shared --print | grep -q 'test/repo3' || exit 13

"$GH_INSTALL_BIN_DIR/githooks-cli" config shared --reset &&
    "$GH_INSTALL_BIN_DIR/githooks-cli" config shared --print | grep -q -i 'none' || exit 14
