#!/usr/bin/env bash
# Test:
#   Cli tool: disable a hook

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

accept_all_trust_prompts || exit 1

"$GH_TEST_BIN/githooks-cli" installer || exit 1

mkdir -p "$GH_TEST_TMP/test056/.githooks/pre-commit" &&
    cd "$GH_TEST_TMP/test056" &&
    echo 'echo "Hello"' >".githooks/pre-commit/first" &&
    echo 'echo "Hello"' >".githooks/pre-commit/second" &&
    echo "test" >".githooks/.namespace" &&
    git init &&
    install_hooks_if_not_centralized || exit 1

if ! "$GH_INSTALL_BIN_DIR/githooks-cli" ignore add --pattern "**/first"; then
    echo "! Failed to disable a hook"
    exit 1
fi

if ! "$GH_INSTALL_BIN_DIR/githooks-cli" list | grep "first" | grep -q "'ignored'" ||
    ! "$GH_INSTALL_BIN_DIR/githooks-cli" list | grep "second" | grep -q "'active'"; then
    echo "! Unexpected cli list output (1)"
    exit 1
fi

if ! "$GH_INSTALL_BIN_DIR/githooks-cli" ignore add --pattern "pre-commit/**"; then
    echo "! Failed to disable a hook"
    exit 1
fi

if ! "$GH_INSTALL_BIN_DIR/githooks-cli" list | grep "first" | grep -q "'ignored'" ||
    ! "$GH_INSTALL_BIN_DIR/githooks-cli" list | grep "second" | grep -q "'ignored'"; then
    echo "! Unexpected cli list output (2)"
    exit 1
fi

# Negate the pattern
if ! "$GH_INSTALL_BIN_DIR/githooks-cli" ignore add --pattern "!**/second"; then
    echo "! Failed to disable a hook"
    exit 1
fi

if ! "$GH_INSTALL_BIN_DIR/githooks-cli" list | grep "first" | grep -q "'ignored'" ||
    ! "$GH_INSTALL_BIN_DIR/githooks-cli" list | grep "second" | grep -q "'active'"; then
    echo "! Unexpected cli list output (3)"
    exit 1
fi

# Negate the pattern more
if ! "$GH_INSTALL_BIN_DIR/githooks-cli" ignore add --pattern "!**/*"; then
    echo "! Failed to disable a hook"
    exit 1
fi

if ! "$GH_INSTALL_BIN_DIR/githooks-cli" list | grep "first" | grep -q "'active'" ||
    ! "$GH_INSTALL_BIN_DIR/githooks-cli" list | grep "second" | grep -q "'active'"; then
    echo "! Unexpected cli list output (4)"
    exit 1
fi

# Exclude all
if ! "$GH_INSTALL_BIN_DIR/githooks-cli" ignore add --pattern "**/*"; then
    echo "! Failed to disable a hook"
    exit 1
fi

if ! "$GH_INSTALL_BIN_DIR/githooks-cli" list | grep "first" | grep -q "'ignored'" ||
    ! "$GH_INSTALL_BIN_DIR/githooks-cli" list | grep "second" | grep -q "'ignored'"; then
    echo "! Unexpected cli list output (5)"
    exit 1
fi

if ! "$GH_INSTALL_BIN_DIR/githooks-cli" ignore remove --all; then
    echo "! Failed to disable alls hooks"
    exit 1
fi

if ! "$GH_INSTALL_BIN_DIR/githooks-cli" list | grep "first" | grep -q "'active'" ||
    ! "$GH_INSTALL_BIN_DIR/githooks-cli" list | grep "second" | grep -q "'active'"; then
    echo "! Unexpected cli list output (6)"
    exit 1
fi

# with full matches by  namespace paths
if ! "$GH_INSTALL_BIN_DIR/githooks-cli" ignore add --pattern "ns:test*/pre-commit/first"; then
    echo "! Failed to disable a hook"
    exit 1
fi

if ! "$GH_INSTALL_BIN_DIR/githooks-cli" list | grep "first" | grep -q "'ignored'" ||
    ! "$GH_INSTALL_BIN_DIR/githooks-cli" list | grep "second" | grep -q "'active'"; then
    echo "! Unexpected cli list output (7)"
    exit 1
fi

if ! "$GH_INSTALL_BIN_DIR/githooks-cli" ignore add --pattern "ns:gh-self/pre-commit/second"; then
    echo "! Failed to disable a hook"
    exit 1
fi

if ! "$GH_INSTALL_BIN_DIR/githooks-cli" list | grep "first" | grep -q "'ignored'" ||
    ! "$GH_INSTALL_BIN_DIR/githooks-cli" list | grep "second" | grep -q "'ignored'"; then
    echo "! Unexpected cli list output (8)"
    exit 1
fi
