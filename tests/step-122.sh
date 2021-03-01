#!/bin/sh
# Test:
#   Check if dialog executable is installed and works

"$GH_TEST_BIN/cli" installer || exit 1

DIALOG=$(git config --global githooks.dialog)

if [ ! -f "$DIALOG" ]; then
    echo "! Dialog tool '$DIALOG' not found." >&2
    exit 2
fi

if ! "$DIALOG" --version >/dev/null 2>&1; then
    echo "! Dialog tool not working properly." >&2
    exit 3
fi

if ! "$DIALOG" --version 2>&1 | grep -q "dialog version"; then
    echo "! Dialog tool not working properly" >&2
    exit 3
fi

if ! "$DIALOG" options --help 2>&1 | grep -q "dialog"; then
    echo "! Dialog tool not working properly" >&2
    exit 3
fi
