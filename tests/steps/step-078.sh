#!/usr/bin/env bash
# Test:
#   Cli tool: enable/disable auto updates

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

init_step

accept_all_trust_prompts || exit 1

if ! "$GH_TEST_BIN/githooks-cli" installer "${EXTRA_INSTALL_ARGS[@]}"; then
    echo "! Failed to execute the install script"
    exit 1
fi

git config --global --unset githooks.updateCheckEnabled &&
    "$GH_INSTALL_BIN_DIR/githooks-cli" update --enable-check &&
    [ "$(git config --get githooks.updateCheckEnabled)" = "true" ] ||
    exit 1

git config --global --unset githooks.updateCheckEnabled &&
    "$GH_INSTALL_BIN_DIR/githooks-cli" update --disable-check &&
    [ "$(git config --get githooks.updateCheckEnabled)" = "false" ] ||
    exit 1
