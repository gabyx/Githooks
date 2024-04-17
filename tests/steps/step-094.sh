#!/usr/bin/env bash
# Test:
#   Cli tool: run an installation
# shellcheck disable=SC1091

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

init_step

accept_all_trust_prompts || exit 1

mkdir -p "$GH_TEST_TMP/test094/a" "$GH_TEST_TMP/test094/b" "$GH_TEST_TMP/test094/c" &&
    cd "$GH_TEST_TMP/test094/a" &&
    git init &&
    cd "$GH_TEST_TMP/test094/b" &&
    git init || exit 1

"$GH_TEST_BIN/githooks-cli" installer "${EXTRA_INSTALL_ARGS[@]}" || exit 1

git config --global githooks.previousSearchDir "$GH_TEST_TMP"

if ! "$GH_INSTALL_BIN_DIR/githooks-cli" installer "${EXTRA_INSTALL_ARGS[@]}"; then
    echo "! Failed to run the global installation"
    exit 1
fi

if is_centralized_tests; then
    check_centralized_install
else
    check_local_install
fi

if (cd "$GH_TEST_TMP/test094/c" && "$GH_INSTALL_BIN_DIR/githooks-cli" install); then
    echo "! Install expected to fail outside a repository"
    exit 1
fi

# Reset to trigger a global update
if ! git -C "$GH_TEST_REPO" reset --hard v9.9.1; then
    echo "! Could not reset server to trigger update."
    exit 1
fi

CURRENT="$(git -C ~/.githooks/release rev-parse HEAD)"
if ! "$GH_INSTALL_BIN_DIR/githooks-cli" installer "${EXTRA_INSTALL_ARGS[@]}"; then
    echo "! Expected global installation to succeed"
    exit 1
fi
AFTER="$(git -C ~/.githooks/release rev-parse HEAD)"
if [ "$CURRENT" = "$AFTER" ] ||
    [ "$(git -C "$GH_TEST_REPO" rev-parse v9.9.1)" != "$AFTER" ]; then
    echo "! Release clone was not updated, but it should have!"
    exit 1
fi
