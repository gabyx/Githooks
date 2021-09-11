#!/bin/sh
# Test:
#   Run a default install and verify the cli helper is installed

TEST_DIR=$(cd "$(dirname "$0")" && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

acceptAllTrustPrompts || exit 1

"$GH_TEST_BIN/cli" installer || exit 1

if ! "$GH_INSTALL_BIN_DIR/cli" --version; then
    echo "! The command line helper tool is not available"
    exit 1
fi
