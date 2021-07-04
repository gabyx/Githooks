#!/bin/sh
# Test:
#   Cli tool: run update check

TEST_DIR=$(cd "$(dirname "$0")" && pwd)
# shellcheck disable=SC1090
. "$TEST_DIR/general.sh"

acceptAllTrustPrompts || exit 1

"$GH_TEST_BIN/cli" installer || exit 1

mkdir -p "$GH_TEST_TMP/test062" &&
    cd "$GH_TEST_TMP/test062" &&
    git init || exit 1

OUT=$("$GH_INSTALL_BIN_DIR/cli" update --no)
# shellcheck disable=SC2181
if [ $? -ne 0 ] || ! echo "$OUT" | grep -qi "is at the latest version"; then
    echo "! Failed to run the update with --no"
    echo "$OUT"
    exit 1
fi

# Reset to trigger update
if ! (cd ~/.githooks/release && git reset --hard v9.9.0 >/dev/null); then
    echo "! Could not reset master to trigger update."
    exit 1
fi

OUT=$("$GH_INSTALL_BIN_DIR/cli" update --no)
# shellcheck disable=SC2181
if [ $? -ne 0 ] || ! echo "$OUT" | grep -qi "update declined"; then
    echo "! Failed to run the update with --no"
    echo "$OUT"
    exit 1
fi
