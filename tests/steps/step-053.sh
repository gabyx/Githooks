#!/usr/bin/env bash
# Test:
#   Cli tool: list current hooks

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

accept_all_trust_prompts || exit 1

"$GH_TEST_BIN/githooks-cli" installer "${EXTRA_INSTALL_ARGS[@]}" || exit 1

mkdir -p "$GH_TEST_TMP/test053/.githooks/pre-commit" &&
    echo 'echo "Hello"' >"$GH_TEST_TMP/test053/.githooks/pre-commit/example" &&
    cd "$GH_TEST_TMP/test053" &&
    git init &&
    install_hooks_if_not_centralized || exit 1

if ! "$GH_INSTALL_BIN_DIR/githooks-cli" list | grep "example" | grep "'untrusted'" | grep "'active'"; then
    echo "! Unexpected cli list output"
    exit 1
fi

git commit -m 'Test'

if ! "$GH_INSTALL_BIN_DIR/githooks-cli" list | grep "example" | grep "'trusted'" | grep "'active'"; then
    echo "! Unexpected cli list output"
    exit 1
fi
