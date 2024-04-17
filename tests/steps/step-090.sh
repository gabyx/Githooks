#!/usr/bin/env bash
# Test:
#   Cli tool: manage disable configuration

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

init_step

accept_all_trust_prompts || exit 1

if ! "$GH_TEST_BIN/githooks-cli" installer "${EXTRA_INSTALL_ARGS[@]}"; then
    echo "! Failed to execute the install script"
    exit 1
fi

mkdir -p "$GH_TEST_TMP/test090" && cd "$GH_TEST_TMP/test090" || exit 2

! "$GH_INSTALL_BIN_DIR/githooks-cli" config disable --set || exit 3 # not a Git repository

git init || exit 4

! "$GH_INSTALL_BIN_DIR/githooks-cli" config disable || exit 5

"$GH_INSTALL_BIN_DIR/githooks-cli" config disable --set &&
    "$GH_INSTALL_BIN_DIR/githooks-cli" config disable --print | grep -q 'is disabled' || exit 6
"$GH_INSTALL_BIN_DIR/githooks-cli" config disable --reset &&
    "$GH_INSTALL_BIN_DIR/githooks-cli" config disable --print | grep -q 'is not disabled' || exit 7
