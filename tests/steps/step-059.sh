#!/usr/bin/env bash
# Test:
#   Cli tool: list shows ignored files

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

init_step

accept_all_trust_prompts || exit 1

"$GH_TEST_BIN/githooks-cli" installer "${EXTRA_INSTALL_ARGS[@]}" || exit 1

mkdir -p "$GH_TEST_TMP/test059/.githooks/pre-commit" &&
    echo 'echo "Hello"' >"$GH_TEST_TMP/test059/.githooks/pre-commit/first" &&
    echo 'echo "Hello"' >"$GH_TEST_TMP/test059/.githooks/pre-commit/second" &&
    echo 'patterns: - pre-commit/first' >"$GH_TEST_TMP/test059/.githooks/.ignore.yaml" &&
    echo 'patterns: - .///second' >"$GH_TEST_TMP/test059/.githooks/pre-commit/.ignore.yaml" &&
    cd "$GH_TEST_TMP/test059" &&
    git init || exit 1

if ! "$GH_INSTALL_BIN_DIR/githooks-cli" list | grep "first" | grep -q "'ignored'"; then
    echo "! Unexpected cli list output (1)"
    exit 1
fi

if ! "$GH_INSTALL_BIN_DIR/githooks-cli" list | grep "second" | grep -q "'ignored'"; then
    echo "! Unexpected cli list output (2)"
    exit 1
fi
