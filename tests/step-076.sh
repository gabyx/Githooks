#!/bin/sh
# Test:
#   Direct runner execution: choose to ignore the update (non-single)

TEST_DIR=$(cd "$(dirname "$0")" && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

mkdir -p "$GH_TEST_TMP/test076" &&
    cd "$GH_TEST_TMP/test076" &&
    git init || exit 1

# Reset to trigger update
git config --global githooks.autoUpdateEnabled true || exit 1

OUTPUT=$(
    ACCEPT_CHANGES=A EXECUTE_UPDATE=N \
        "$GH_TEST_BIN/runner" "$(pwd)"/.git/hooks/post-commit 2>&1
)

if ! git -C ~/.githooks/release rev-parse HEAD; then
    echo "! Release clone was not cloned, but it should have!"
    exit 1
fi

LAST_UPDATE=$(git config --global --get githooks.autoUpdateCheckTimestamp)
if [ -z "$LAST_UPDATE" ]; then
    echo "! Update was expected to start"
    exit 1
fi

# 'git' is removed in 'hooks update disable'
# due to covarage replacement issues.
if ! echo "$OUTPUT" | grep -q "hooks update disable"; then
    echo "! Expected update output not found"
    echo "$OUTPUT"
    exit 1
fi
