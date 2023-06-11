#!/usr/bin/env bash
# Test:
#   Run a single-repo, non-interactive install successfully

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

acceptAllTrustPrompts || exit 1

if echo "$EXTRA_INSTALL_ARGS" | grep -q "use-core-hookspath"; then
    echo "Using core.hooksPath"
    exit 249
fi

mkdir -p "$GH_TEST_TMP/start/dir" &&
    cd "$GH_TEST_TMP/start/dir" &&
    git init || exit 1

if ! "$GH_TEST_BIN/cli" installer --non-interactive; then
    echo "! Installation failed"
    exit 1
fi

if ! "$GH_TEST_BIN/cli" install --non-interactive; then
    echo "! Install into current repo failed"
    exit 1
fi

if ! grep -r 'github.com/gabyx/githooks' "$GH_TEST_TMP/start/dir/.git/hooks"; then
    echo "! Hooks were not installed"
    exit 1
fi
