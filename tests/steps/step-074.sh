#!/usr/bin/env bash
# Test:
#   Cli tool: list pending changes

if ! "$GH_TEST_BIN/githooks-cli" installer "${EXTRA_INSTALL_ARGS[@]}"; then
    echo "! Failed to execute the install script"
    exit 1
fi

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

accept_all_trust_prompts || exit 1

mkdir -p "$GH_TEST_TMP/test074/.githooks/pre-commit" &&
    echo 'echo "Hello"' >"$GH_TEST_TMP/test074/.githooks/pre-commit/testing" &&
    cd "$GH_TEST_TMP/test074" &&
    git init || exit 1

if ! "$GH_INSTALL_BIN_DIR/githooks-cli" list pre-commit | grep 'testing' | grep "'active'" | grep -q "'untrusted'"; then
    echo "! Unexpected list result (1)"
    exit 1
fi

if ! "$GH_INSTALL_BIN_DIR/githooks-cli" trust hooks --path "pre-commit/testing"; then
    echo "! Failed to accept the hook"
    exit 1
fi

if ! "$GH_INSTALL_BIN_DIR/githooks-cli" list pre-commit | grep 'testing' | grep "'active'" | grep -q "'trusted'"; then
    echo "! Unexpected list result (2)"
    exit 1
fi

echo 'echo "Changed"' >"$GH_TEST_TMP/test074/.githooks/pre-commit/testing" || exit 1

if ! "$GH_INSTALL_BIN_DIR/githooks-cli" list pre-commit | grep 'testing' | grep "'active'" | grep -q "'untrusted'"; then
    echo "! Unexpected list result (2)"
    exit 1
fi
