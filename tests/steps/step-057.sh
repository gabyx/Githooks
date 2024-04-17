#!/usr/bin/env bash
# Test:
#   Cli tool: enable a hook

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

accept_all_trust_prompts || exit 1

"$GH_TEST_BIN/githooks-cli" installer "${EXTRA_INSTALL_ARGS[@]}" || exit 1

mkdir -p "$GH_TEST_TMP/test057/.githooks/pre-commit" &&
    echo 'echo "Hello"' >"$GH_TEST_TMP/test057/.githooks/pre-commit/first" &&
    echo 'echo "Hello"' >"$GH_TEST_TMP/test057/.githooks/pre-commit/second" &&
    cd "$GH_TEST_TMP/test057" &&
    git init || exit 1

if ! "$GH_INSTALL_BIN_DIR/githooks-cli" ignore add --pattern "**/*"; then
    echo "! Failed ignore hooks"
    exit 1
fi

if ! "$GH_INSTALL_BIN_DIR/githooks-cli" list | grep "first" | grep -q "'ignored'" ||
    ! "$GH_INSTALL_BIN_DIR/githooks-cli" list | grep "second" | grep -q "'ignored'"; then
    echo "! Unexpected cli list output (1)"
    exit 1
fi

if ! "$GH_INSTALL_BIN_DIR/githooks-cli" disable; then
    echo "! Failed to disable Githooks"
    exit 1
fi

if ! "$GH_INSTALL_BIN_DIR/githooks-cli" list | grep "first" | grep -q "'disabled'" ||
    ! "$GH_INSTALL_BIN_DIR/githooks-cli" list | grep "second" | grep -q "'disabled'"; then
    echo "! Unexpected cli list output (2)"
    exit 1
fi

if ! "$GH_INSTALL_BIN_DIR/githooks-cli" ignore add --pattern '!**/*'; then
    echo "! Failed to ignore hooks"
    exit 1
fi

if ! "$GH_INSTALL_BIN_DIR/githooks-cli" list | grep "first" | grep -q "'disabled'" ||
    ! "$GH_INSTALL_BIN_DIR/githooks-cli" list | grep "second" | grep -q "'disabled'"; then
    echo "! Unexpected cli list output (3)"
    exit 1
fi

if ! "$GH_INSTALL_BIN_DIR/githooks-cli" disable --reset; then
    echo "! Failed to reset disabling Githooks"
    exit 1
fi

if ! "$GH_INSTALL_BIN_DIR/githooks-cli" list | grep "first" | grep -q "'active'" ||
    ! "$GH_INSTALL_BIN_DIR/githooks-cli" list | grep "second" | grep -q "'active'"; then
    echo "! Unexpected cli list output (4)"
    "$GH_INSTALL_BIN_DIR/githooks-cli" list
    exit 1
fi

if ! "$GH_INSTALL_BIN_DIR/githooks-cli" ignore remove --pattern "**/*"; then
    echo "! Failed to remove a pattern"
    exit 1
fi

if ! "$GH_INSTALL_BIN_DIR/githooks-cli" ignore add --pattern "**/*"; then
    echo "! Failed to add a pattern back to the end of the list"
    exit 1
fi

if ! "$GH_INSTALL_BIN_DIR/githooks-cli" list | grep "first" | grep -q "'ignored'" ||
    ! "$GH_INSTALL_BIN_DIR/githooks-cli" list | grep "second" | grep -q "'ignored'"; then
    echo "! Unexpected cli list output (5)"
    exit 1
fi
