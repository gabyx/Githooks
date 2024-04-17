#!/usr/bin/env bash
# Test:
#   Check if dialog executable is installed and works

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

init_step

accept_all_trust_prompts || exit 1

"$GH_TEST_BIN/githooks-cli" installer "${EXTRA_INSTALL_ARGS[@]}" || exit 1

DIALOG=$(git config --global githooks.dialog)

if [ ! -f "$DIALOG" ]; then
    echo "! Dialog tool '$DIALOG' not found."
    exit 2
fi

if ! "$GH_INSTALL_BIN_DIR/githooks-dialog" --version >/dev/null 2>&1; then
    echo "! Dialog tool not working properly."
    exit 3
fi

if ! "$GH_INSTALL_BIN_DIR/githooks-dialog" --version 2>&1 | grep -q "dialog version"; then
    echo "! Dialog tool not working properly"
    exit 4
fi

if ! "$GH_INSTALL_BIN_DIR/githooks-dialog" options --help 2>&1 | grep -q "dialog"; then
    echo "! Dialog tool not working properly"
    exit 5
fi
