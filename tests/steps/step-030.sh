#!/usr/bin/env bash
# Test:
#   Direct runner execution: auto-update is not enabled

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

mkdir -p "$GH_TEST_TMP/test30" &&
    cd "$GH_TEST_TMP/test30" &&
    git init || exit 1

git config --global githooks.autoUpdateEnabled false || exit 1

ACCEPT_CHANGES=A "$GH_TEST_BIN/runner" "$(pwd)"/.git/hooks/post-commit

# shellcheck disable=SC2181
if git -C ~/.githooks/release rev-parse HEAD; then
    echo "! Release clone cloned, but it should not have!"
    exit 1
fi

LAST_UPDATE=$(get_update_check_timestamp)
if [ -n "$LAST_UPDATE" ]; then
    echo "! Update unexpectedly run"
    exit 1
fi
