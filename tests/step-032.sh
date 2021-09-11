#!/usr/bin/env bash
# Test:
#   Direct runner execution: choose to ignore the update

TEST_DIR=$(cd "$(dirname "$0")" && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

mkdir -p "$GH_TEST_TMP/test32" &&
    cd "$GH_TEST_TMP/test32" &&
    git init || exit 1

git config --global githooks.autoUpdateEnabled true || exit 1

OUTPUT=$(
    ACCEPT_CHANGES=A EXECUTE_UPDATE=N \
        "$GH_TEST_BIN/runner" "$(pwd)"/.git/hooks/post-commit 2>&1
)

if ! git -C ~/.githooks/release rev-parse HEAD; then
    echo "! Release clone was not updated, but it should have!"
    exit 1
fi

LAST_UPDATE=$(git config --global --get githooks.autoUpdateCheckTimestamp)
if [ -z "$LAST_UPDATE" ]; then
    echo "! Update check was expected to start"
    exit 1
fi

if ! echo "$OUTPUT" | grep -q "If you would like to disable auto-updates"; then
    echo "! Expected update output not found"
    echo "$OUTPUT"
    exit 1
fi
