#!/bin/sh
# Test:
#   Cli tool: run an update

TEST_DIR=$(cd "$(dirname "$0")" && pwd)
# shellcheck disable=SC1090
. "$TEST_DIR/general.sh"

acceptAllTrustPrompts || exit 1

"$GH_TEST_BIN/cli" installer || exit 1

mkdir -p "$GH_TEST_TMP/test063" &&
    cd "$GH_TEST_TMP/test063" &&
    git init || exit 1

# Reset to trigger update
if ! (cd ~/.githooks/release && git reset --hard v9.9.0 >/dev/null); then
    echo "! Could not reset master to trigger update."
    exit 1
fi

# Update to version 9.9.1
CURRENT="$(cd ~/.githooks/release && git rev-parse HEAD)"
if ! "$GH_INSTALL_BIN_DIR/cli" update --yes; then
    echo "! Failed to run the update"
fi
AFTER="$(cd ~/.githooks/release && git rev-parse HEAD)"

if [ "$CURRENT" = "$AFTER" ]; then
    echo "! Release clone was not updated, but it should have!"
    exit 1
fi

# Update to version 10.1.1
# input "enter"
CURRENT="$AFTER"
out=$(EXECUTE_UPDATE="" "$GH_INSTALL_BIN_DIR/cli" update 2>&1) || {
    echo "! Failed to run update"
    echo "$out"
    exit 1
}
AFTER2="$(cd ~/.githooks/release && git rev-parse HEAD)"

if [ "$CURRENT" != "$AFTER2" ]; then
    echo "! Release clone was updated, but it should not have!"
    echo "$out"
    exit 1
fi

if ! echo "$out" | grep -q -E "Would you like to install.*[y/N]"; then
    echo "! Expected default update answer to be 'no'."
    echo "$out"
    exit 1
fi

# Try again, but now force it.
CURRENT="$AFTER"
out=$("$GH_INSTALL_BIN_DIR/cli" update --yes 2>&1) || {
    echo "! Failed to run update"
    echo "$out"
    exit 1
}
AFTER2="$(cd ~/.githooks/release && git rev-parse HEAD)"

if [ "$CURRENT" = "$AFTER2" ]; then
    echo "! Release clone was not updated, but it should have!"
    echo "$out"
    exit 1
fi

if ! echo "$out" | grep -q "Update Info:" ||
    ! echo "$out" | grep -q "Bug fixes and improvements." ||
    ! echo "$out" | grep -q "Breaking changes for v10.x.x, read the change log."; then
    echo "! Expected update info to be present in output."
    echo "$out"
    exit 1
fi
