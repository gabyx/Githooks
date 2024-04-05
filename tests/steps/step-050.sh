#!/usr/bin/env bash
# Test:
#   Run a default install and verify the cli helper is installed

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

accept_all_trust_prompts || exit 1

"$GH_TEST_BIN/githooks-cli" installer || exit 1

if ! "$GH_INSTALL_BIN_DIR/githooks-cli" --version; then
    echo "! The command line helper tool is not available"
    exit 1
fi

if [ -z "$GH_ON_WINDOWS" ]; then
    find "/tmp" -type f -name "githooks-installer*"
    nCount=$(find "/tmp" -type f -name "githooks-installer-*.log" | wc -l)
    if [ "$nCount" != "1" ]; then
        echo "! The installer log should be created."
        exit 1
    fi
fi
