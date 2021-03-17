#!/bin/sh
# Test:
#   Direct runner execution: execute auto-update

TEST_DIR=$(cd "$(dirname "$0")" && pwd)
# shellcheck disable=SC1090
. "$TEST_DIR/general.sh"

LAST_UPDATE=$(git config --global --get githooks.autoUpdateCheckTimestamp)
if [ -n "$LAST_UPDATE" ]; then
    echo "! Update already marked as run"
    exit 1
fi

mkdir -p "$GH_TEST_TMP/test29" &&
    cd "$GH_TEST_TMP/test29" &&
    git init || exit 1

git config --global githooks.autoUpdateEnabled true || exit 1

ACCEPT_CHANGES=A "$GH_TEST_BIN/runner" "$(pwd)"/.git/hooks/post-commit

if ! cd ~/.githooks/release && git rev-parse HEAD; then
    echo "! Release clone was not updated, but it should have!"
    exit 1
fi

LAST_UPDATE=$(git config --global --get githooks.autoUpdateCheckTimestamp)
if [ -z "$LAST_UPDATE" ]; then
    echo "! Update check did not run"
    exit 1
fi

CURRENT_TIME=$(date +%s)
ELAPSED_TIME=$((CURRENT_TIME - LAST_UPDATE))

if [ $ELAPSED_TIME -gt 5 ]; then
    echo "! Update did not execute properly"
    exit 1
fi
