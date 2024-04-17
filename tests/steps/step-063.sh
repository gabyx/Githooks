#!/usr/bin/env bash
# Test:
#   Cli tool: run an update
# shellcheck disable=SC1091

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

init_step

accept_all_trust_prompts || exit 1

"$GH_TEST_BIN/githooks-cli" installer "${EXTRA_INSTALL_ARGS[@]}" || exit 1

mkdir -p "$GH_TEST_TMP/test063" &&
    cd "$GH_TEST_TMP/test063" &&
    git init || exit 1

# Reset to trigger update
if ! git -C "$GH_TEST_REPO" reset --hard v9.9.1 >/dev/null; then
    echo "! Could not reset server to trigger update."
    exit 1
fi

# Update to version 9.9.1
echo "Update to version 9.9.1"
CURRENT="$(git -C ~/.githooks/release rev-parse HEAD)"
if ! "$GH_INSTALL_BIN_DIR/githooks-cli" update --yes; then
    echo "! Failed to run the update"
fi
AFTER="$(git -C ~/.githooks/release rev-parse HEAD)"

if [ "$CURRENT" = "$AFTER" ] ||
    [ "$(git -C "$GH_TEST_REPO" rev-parse v9.9.1)" != "$AFTER" ]; then
    echo "! Release clone was not updated, but it should have!"
    exit 1
fi

# Reset to trigger update
if ! git -C "$GH_TEST_REPO" reset --hard v10.1.1 >/dev/null; then
    echo "! Could not reset server to trigger update."
    exit 1
fi

echo "Try update to 10.1.1 (1)"
# Update to version 10.1.1
# input "enter"
CURRENT="$AFTER"
out=$(EXECUTE_UPDATE="" "$GH_INSTALL_BIN_DIR/githooks-cli" update 2>&1) || {
    echo "! Failed to run update"
    exit 1
}
AFTER="$(git -C ~/.githooks/release rev-parse HEAD)"

if [ "$CURRENT" != "$AFTER" ] ||
    [ "$(git -C "$GH_TEST_REPO" rev-parse v9.9.1)" != "$AFTER" ]; then
    echo "$out"
    echo "! Release clone was updated, but it should not have!"
    exit 1
fi

if ! echo "$out" | grep -q -E "Would you like to install.*[y/N]"; then
    echo "$out"
    echo "! Expected default update answer to be 'no'."
    exit 1
fi

# Update to version 10.1.1 (its a major update which should be declined)
echo "Try update to 10.1.1 (2)"
CURRENT="$AFTER"
out=$(EXECUTE_UPDATE="" "$GH_INSTALL_BIN_DIR/githooks-cli" update --yes 2>&1) || {
    echo "$out"
    echo "! Failed to run update"
    exit 1
}
AFTER="$(git -C ~/.githooks/release rev-parse HEAD)"

if [ "$CURRENT" != "$AFTER" ] ||
    [ "$(git -C "$GH_TEST_REPO" rev-parse v9.9.1)" != "$AFTER" ]; then
    echo "$out"
    echo "! Release clone was updated, but it should not have!"
    exit 1
fi

# Try again, but now force the major update.
echo "Force update to 10.1.1"
CURRENT="$AFTER"
out=$("$GH_INSTALL_BIN_DIR/githooks-cli" update --yes-all 2>&1) || {
    echo "$out"
    echo "! Failed to run update"
    exit 1
}
AFTER="$(git -C ~/.githooks/release rev-parse HEAD)"

if [ "$CURRENT" = "$AFTER" ] ||
    [ "$(git -C "$GH_TEST_REPO" rev-parse v10.1.1)" != "$AFTER" ]; then
    echo "$out"
    echo "! Release clone was not updated, but it should have!"
    exit 1
fi

if ! echo "$out" | grep -q "Update Info:" ||
    ! echo "$out" | grep -q "Bug fixes and improvements." ||
    ! echo "$out" | grep -q "Breaking changes, read the change log."; then
    echo "$out"
    echo "! Expected update info to be present in output."
    exit 1
fi

echo "Update to pre-release 10.1.2-rc1"
# Reset to trigger update
if ! git -C "$GH_TEST_REPO" reset --hard v10.1.2-rc1 >/dev/null; then
    echo "! Could not reset server to trigger update."
    exit 1
fi

echo "Try update to 10.1.2-rc1"
CURRENT="$AFTER"
out=$(EXECUTE_UPDATE="" "$GH_INSTALL_BIN_DIR/githooks-cli" update --yes-all 2>&1) || {
    echo "$out"
    echo "! Failed to run update"
    exit 1
}
AFTER="$(git -C ~/.githooks/release rev-parse HEAD)"

if [ "$CURRENT" != "$AFTER" ] ||
    [ "$(git -C "$GH_TEST_REPO" rev-parse v10.1.1)" != "$AFTER" ]; then
    echo "$out"
    echo "! Release clone was updated, but it should not have!"
    exit 1
fi

echo "Force update to 10.1.2-rc1"
CURRENT="$AFTER"
out=$(EXECUTE_UPDATE="" "$GH_INSTALL_BIN_DIR/githooks-cli" update --yes-all --use-pre-release 2>&1) || {
    echo "$out"
    echo "! Failed to run update"
    exit 1
}
AFTER="$(git -C ~/.githooks/release rev-parse HEAD)"

if [ "$CURRENT" = "$AFTER" ] ||
    [ "$(git -C "$GH_TEST_REPO" rev-parse v10.1.2-rc1)" != "$AFTER" ]; then
    echo "$out"
    echo "! Release clone was not updated, but it should have!"
    exit 1
fi
