#!/usr/bin/env bash
# shellcheck disable=SC1091
# Test:
#   Run a single-repo install and try the auto-update

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

acceptAllTrustPrompts || exit 1

if echo "${EXTRA_INSTALL_ARGS:-}" | grep -q "use-core-hookspath"; then
    echo "Using core.hooksPath"
    exit 249
fi

LAST_UPDATE=$(getUpdateCheckTimestamp)
if [ -n "$LAST_UPDATE" ]; then
    echo "! Update already marked as run"
    exit 1
fi

mkdir -p "$GH_TEST_TMP/start/dir" &&
    cd "$GH_TEST_TMP/start/dir" &&
    git init || exit 1

if ! "$GH_TEST_BIN/cli" installer; then
    echo "! Installation failed"
    exit 1
fi

if ! "$GH_TEST_BIN/cli" install; then
    echo "! Install into current repo failed"
    exit 1
fi

ARE_UPDATES_ENABLED=$(git config --global --get githooks.autoUpdateEnabled)
if [ "$ARE_UPDATES_ENABLED" != "true" ]; then
    echo "! Auto updates were expected to be enabled"
    exit 1
fi

LAST_UPDATE=$(getUpdateCheckTimestamp)
if [ -n "$LAST_UPDATE" ]; then
    echo "! Update already marked as run"
    exit 1
fi

# Reset to trigger update
if ! git -C "$GH_TEST_REPO" reset --hard v9.9.1 >/dev/null; then
    echo "! Could not reset server to trigger update."
    exit 1
fi

resetUpdateCheckTimestamp

OUTPUT=$(
    "$GH_INSTALL_BIN_DIR/runner" "$(pwd)"/.git/hooks/post-commit 2>&1
)

if ! echo "$OUTPUT" | grep -q "All done! Enjoy!"; then
    echo "! Expected installation output not found"
    echo "$OUTPUT"
    exit 1
fi

LAST_UPDATE=$(getUpdateCheckTimestamp)
if [ -z "$LAST_UPDATE" ]; then
    echo "! Update did not run"
    exit 1
fi

CURRENT_TIME=$(date +%s)
ELAPSED_TIME=$((CURRENT_TIME - LAST_UPDATE))

if [ $ELAPSED_TIME -eq 0 ]; then
    echo "! Update did not execute properly"
    exit 1
fi
