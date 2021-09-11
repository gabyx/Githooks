#!/bin/sh
# shellcheck disable=SC1091
# Test:
#   Cli tool: run an installation

TEST_DIR=$(cd "$(dirname "$0")" && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

acceptAllTrustPrompts || exit 1

mkdir -p "$GH_TEST_TMP/test094/a" "$GH_TEST_TMP/test094/b" "$GH_TEST_TMP/test094/c" &&
    cd "$GH_TEST_TMP/test094/a" &&
    git init &&
    cd "$GH_TEST_TMP/test094/b" &&
    git init || exit 1

"$GH_TEST_BIN/cli" installer || exit 1

git config --global githooks.previousSearchDir "$GH_TEST_TMP"

if ! "$GH_INSTALL_BIN_DIR/cli" installer; then
    echo "! Failed to run the global installation"
    exit 1
fi

if echo "$EXTRA_INSTALL_ARGS" | grep -q "use-core-hookspath"; then
    if [ -f "$GH_TEST_TMP/test094/a/.git/hooks/pre-commit" ]; then
        echo "! Expected hooks not installed"
        exit 1
    fi
else
    if ! grep 'gabyx/githooks' "$GH_TEST_TMP/test094/a/.git/hooks/pre-commit"; then
        echo "! Expected hooks installed"
        exit 1
    fi
fi

if (cd "$GH_TEST_TMP/test094/c" && "$GH_INSTALL_BIN_DIR/cli" install); then
    echo "! Install expected to fail outside a repository"
    exit 1
fi

# Reset to trigger a global update
if ! git -C "$GH_TEST_REPO" reset --hard v9.9.1; then
    echo "! Could not reset server to trigger update."
    exit 1
fi

CURRENT="$(git -C ~/.githooks/release rev-parse HEAD)"
if ! "$GH_INSTALL_BIN_DIR/cli" installer; then
    echo "! Expected global installation to succeed"
    exit 1
fi
AFTER="$(git -C ~/.githooks/release rev-parse HEAD)"
if [ "$CURRENT" = "$AFTER" ] ||
    [ "$(git -C "$GH_TEST_REPO" rev-parse v9.9.1)" != "$AFTER" ]; then
    echo "! Release clone was not updated, but it should have!"
    exit 1
fi
