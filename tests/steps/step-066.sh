#!/usr/bin/env bash
# Test:
#   Run an install that tries to install hooks into a non-existing directory

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

init_step

accept_all_trust_prompts || exit 1

if echo "${EXTRA_INSTALL_ARGS:-}" | grep -q "centralized"; then
    echo "Using centralized install"
    exit 249
fi

rm -rf /does/not/exist

OUTPUT=$(
    echo 'y

n
y
/does/not/exist
' | "$GH_TEST_BIN/githooks-cli" installer "${EXTRA_INSTALL_ARGS[@]}" --stdin 2>&1
)

if ! echo "$OUTPUT" | grep "Answer must be an existing directory"; then
    echo "$OUTPUT"
    echo "! Expected error message not found"
    exit 1
fi
