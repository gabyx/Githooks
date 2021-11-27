#!/usr/bin/env bash
# Test:
#   Direct runner execution: accept changes to hooks

TEST_DIR=$(cd "$(dirname "$0")" && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

if [ -n "$GH_ON_WINDOWS" ]; then
    echo "On Windows."
    exit 249
fi

"$TEST_DIR/step-028.sh" --use-symbolic-link
