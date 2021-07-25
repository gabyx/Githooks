#!/bin/sh
# shellcheck disable=SC1091
# Test:
#   Cli tool: run an update by building from source

TEST_DIR=$(cd "$(dirname "$0")" && pwd)
# shellcheck disable=SC1090
. "$TEST_DIR/general.sh"

acceptAllTrustPrompts || exit 1

"$GH_TEST_BIN/cli" installer || exit 1

mkdir -p "$GH_TEST_TMP/test064" &&
    cd "$GH_TEST_TMP/test064" &&
    git init || exit 1

# Reset to trigger update
if ! git -C "$GH_TEST_REPO" reset --hard v9.9.1 >/dev/null; then
    echo "! Could not reset server to trigger update."
    exit 1
fi

# Set to build from source
git config --global githooks.buildFromSource "true"

CURRENT="$(git -C ~/.githooks/release rev-parse HEAD)"
if ! OUT=$("$GH_INSTALL_BIN_DIR/cli" update --yes); then
    echo "! Failed to run the update"
fi

if ! echo "$OUT" | grep -qi "building from source"; then
    echo "! Did not build from source."
    exit 1
fi

AFTER="$(git -C ~/.githooks/release rev-parse HEAD)"

if [ "$CURRENT" = "$AFTER" ] ||
    [ "$(git -C "$GH_TEST_REPO" rev-parse v9.9.1)" != "$AFTER" ]; then
    echo "! Release clone was not updated, but it should have!"
    exit 1
fi

if ! "$GH_INSTALL_BIN_DIR/cli" --version | grep -q "9.9.1"; then
    echo "! Expected to update to 9.9.1"
    "$GH_INSTALL_BIN_DIR/cli" --version
    exit 1
fi

# Check that current commit has `Update-NoSkip: true` trailer
if ! git -C ~/.githooks/release log -n 1 "$AFTER" --pretty="%(trailers:key=Update-NoSkip,valueonly)" |
    grep -q "true"; then
    echo "! Did not detect 'Update-NoSkip: true' trailer on current release commit."
    exit 1
fi
