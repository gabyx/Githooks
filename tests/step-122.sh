#!/bin/sh
# Test:
#   Check if dialog executable is installed and works

TEST_DIR=$(cd "$(dirname "$0")" && pwd)
# shellcheck disable=SC1090
. "$TEST_DIR/general.sh"

acceptAllTrustPrompts || exit 1

"$GH_TEST_BIN/cli" installer || exit 1

DIALOG=$(git config --global githooks.dialog)

if [ ! -f "$DIALOG" ]; then
    echo "! Dialog tool '$DIALOG' not found."
    exit 2
fi

if ! "$GITHOOKS_INSTALL_BIN_DIR/dialog" --version >/dev/null 2>&1; then
    echo "! Dialog tool not working properly."
    exit 3
fi

if ! "$GITHOOKS_INSTALL_BIN_DIR/dialog" --version 2>&1 | grep -q "dialog version"; then
    echo "! Dialog tool not working properly"
    exit 4
fi

if ! "$GITHOOKS_INSTALL_BIN_DIR/dialog" options --help 2>&1 | grep -q "dialog"; then
    echo "! Dialog tool not working properly"
    exit 5
fi
