#!/usr/bin/env bash
# Test:
#   Run the cli tool trying to list hooks of invalid type

if ! "$GH_TEST_BIN/githooks-cli" installer "${EXTRA_INSTALL_ARGS[@]}"; then
    echo "! Failed to execute the install script"
    exit 1
fi

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

accept_all_trust_prompts || exit 1

mkdir -p "$GH_TEST_TMP/test072/.githooks/pre-commit" &&
    echo 'echo "Hello"' >"$GH_TEST_TMP/test072/.githooks/pre-commit/testing" &&
    cd "$GH_TEST_TMP/test072" &&
    git init || exit 1

if ! "$GH_INSTALL_BIN_DIR/githooks-cli" list pre-commit; then
    echo "! Failed to execute a valid list"
    exit 1
fi

if ! "$GH_INSTALL_BIN_DIR/githooks-cli" list invalid-type 2>&1 | grep -q 'not managed by'; then
    echo "! Unexpected list result"
    exit 1
fi
