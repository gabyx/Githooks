#!/usr/bin/env bash
# Test:
#   Run a non-interactive install successfully

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

accept_all_trust_prompts || exit 1

if echo "${EXTRA_INSTALL_ARGS:-}" | grep -q "centralized"; then
    echo "Using centralized install"
    exit 249
fi

mkdir -p "$GH_TEST_TMP/start/dir" &&
    cd "$GH_TEST_TMP/start/dir" &&
    git init || exit 1

if ! "$GH_TEST_BIN/githooks-cli" installer --non-interactive; then
    echo "! Installation failed"
    exit 1
fi

# Install
if ! "$GH_TEST_BIN/githooks-cli" install --non-interactive; then
    echo "! Install into current repo failed"
    exit 1
fi

check_local_install

# Uninstall
if ! "$GH_TEST_BIN/githooks-cli" uninstall; then
    echo "! Uninstall into current repo failed"
    exit 1
fi

check_no_local_install
