#!/usr/bin/env bash
# Test:
#   Cli tool: accept changes to a hook

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

accept_all_trust_prompts || exit 1

"$GH_TEST_BIN/cli" installer || exit 1

mkdir -p "$GH_TEST_TMP/test058/.githooks/pre-commit" &&
    echo 'echo "Hello 1"' >"$GH_TEST_TMP/test058/.githooks/pre-commit/first" &&
    echo 'echo "Hello 2"' >"$GH_TEST_TMP/test058/.githooks/pre-commit/second" &&
    cd "$GH_TEST_TMP/test058" &&
    git init || exit 1

if ! "$GH_INSTALL_BIN_DIR/cli" list | grep "first" | grep -q "'untrusted'" ||
    ! "$GH_INSTALL_BIN_DIR/cli" list | grep "second" | grep -q "'untrusted'"; then
    echo "! Unexpected cli list output (1)"
    exit 1
fi

if ! "$GH_INSTALL_BIN_DIR/cli" trust hooks --pattern pre-commit/first; then
    echo "! Failed to accept a hook by relative path"
    exit 1
fi

if ! "$GH_INSTALL_BIN_DIR/cli" list | grep "first" | grep -q "'trusted'" ||
    ! "$GH_INSTALL_BIN_DIR/cli" list | grep "second" | grep -q "'untrusted'"; then
    echo "! Unexpected cli list output (2)"
    "$GH_INSTALL_BIN_DIR/cli" list
    exit 1
fi

if ! "$GH_INSTALL_BIN_DIR/cli" trust hooks --pattern "**/*"; then
    echo "! Failed to accept all hooks"
    exit 1
fi

if ! "$GH_INSTALL_BIN_DIR/cli" list | grep "first" | grep -q "'trusted'" ||
    ! "$GH_INSTALL_BIN_DIR/cli" list | grep "second" | grep -q "'trusted'"; then
    echo "! Unexpected cli list output (3)"
    exit 1
fi

if ! "$GH_INSTALL_BIN_DIR/cli" trust hooks --reset --pattern "**/*"; then
    echo "! Failed to accept all hooks"
    exit 1
fi

if ! "$GH_INSTALL_BIN_DIR/cli" list | grep "first" | grep -q "'untrusted'" ||
    ! "$GH_INSTALL_BIN_DIR/cli" list | grep "second" | grep -q "'untrusted'"; then
    echo "! Unexpected cli list output (4)"
    exit 1
fi

if ! "$GH_INSTALL_BIN_DIR/cli" trust hooks --all; then
    echo "! Failed to accept all hooks"
    exit 1
fi

if ! "$GH_INSTALL_BIN_DIR/cli" list | grep "first" | grep -q "'trusted'" ||
    ! "$GH_INSTALL_BIN_DIR/cli" list | grep "second" | grep -q "'trusted'"; then
    echo "! Unexpected cli list output (5)"
    exit 1
fi

if ! "$GH_INSTALL_BIN_DIR/cli" trust hooks --reset --all; then
    echo "! Failed to accept all hooks"
    exit 1
fi

if ! "$GH_INSTALL_BIN_DIR/cli" list | grep "first" | grep -q "'untrusted'" ||
    ! "$GH_INSTALL_BIN_DIR/cli" list | grep "second" | grep -q "'untrusted'"; then
    echo "! Unexpected cli list output (6)"
    exit 1
fi
