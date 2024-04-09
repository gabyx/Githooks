#!/usr/bin/env bash
# Test:
#   Run a single-repo, dry-run install successfully

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

if ! "$GH_TEST_BIN/githooks-cli" installer --dry-run; then
    echo "! Installation failed"
    exit 1
fi

if grep -r 'github.com/gabyx/githooks' "$GH_TEST_TMP/start/dir/.git/hooks"; then
    echo "! Hooks were not expected to be installed"
    exit 1
fi
