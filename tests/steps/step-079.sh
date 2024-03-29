#!/usr/bin/env bash
# Test:
#   Cli tool: enable/disable hooks

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

acceptAllTrustPrompts || exit 1

if ! "$GH_TEST_BIN/cli" installer; then
    echo "! Failed to execute the install script"
    exit 1
fi

mkdir -p "$GH_TEST_TMP/test079" &&
    cd "$GH_TEST_TMP/test079" &&
    git init || exit 1

"$GH_INSTALL_BIN_DIR/cli" disable &&
    [ "$(git config --get githooks.disable)" = "true" ] ||
    exit 1

"$GH_INSTALL_BIN_DIR/cli" disable --reset &&
    [ "$(git config --get githooks.disable)" != "true" ] ||
    exit 1
