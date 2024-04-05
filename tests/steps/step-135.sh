#!/usr/bin/env bash
# Test:
#   Run config enable-containerized-hooks
set -e
set -u

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091

. "$TEST_DIR/general.sh"

"$GH_TEST_BIN/githooks-cli" installer || exit 1

mkdir -p "$GH_TEST_TMP/test135" &&
    cd "$GH_TEST_TMP/test135" &&
    git init

# Enable
"$GH_TEST_BIN/githooks-cli" config enable-containerized-hooks --local --set
if [ "$(git config --local githooks.containerizedHooksEnabled)" != "true" ]; then
    echo "Containerized not enabled."
    exit 1
fi

"$GH_TEST_BIN/githooks-cli" config enable-containerized-hooks --print |
    grep -iE ".*containerized.*enabled locally" ||
    die "Failed to enable"

"$GH_TEST_BIN/githooks-cli" config enable-containerized-hooks --global --set
if [ "$(git config --global githooks.containerizedHooksEnabled)" != "true" ]; then
    echo "Containerized not enabled."
    exit 1
fi

"$GH_TEST_BIN/githooks-cli" config enable-containerized-hooks --global --print |
    grep -iE ".*containerized.*enabled globally" ||
    die "Failed to enable"

# Disable
"$GH_TEST_BIN/githooks-cli" config enable-containerized-hooks --global --reset
if [ "$(git config --global githooks.containerizedHooksEnabled)" != "" ]; then
    echo "Containerized not disabled."
    exit 1
fi

"$GH_TEST_BIN/githooks-cli" config enable-containerized-hooks --global --print |
    grep -iE ".*containerized.*disabled globally" ||
    die "Failed to disable"

"$GH_TEST_BIN/githooks-cli" config enable-containerized-hooks --local --reset
if [ "$(git config --local githooks.containerizedHooksEnabled)" != "" ]; then
    echo "Containerized not disabled."
    exit 1
fi

"$GH_TEST_BIN/githooks-cli" config enable-containerized-hooks --local --print |
    grep -iE ".*containerized.*disabled locally" ||
    die "Failed to disable"
