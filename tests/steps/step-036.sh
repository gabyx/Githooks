#!/usr/bin/env bash
# Test:
#   Automatic update checks are already enabled

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

accept_all_trust_prompts || exit 1

echo 'y
' | "$GH_TEST_BIN/githooks-cli" installer "${EXTRA_INSTALL_ARGS[@]}" --stdin || exit 1

if [ "$(git config --global --get githooks.updateCheckEnabled)" != "true" ]; then
    echo "! Automatic update checks are not enabled"
    exit 1
fi

path=$(git config --global githooks.pathForUseCoreHooksPath)
[ -d "$path" ] || {
    echo "! Path '$path' does not exist."
    exit 1
}

OUTPUT=$("$GH_TEST_BIN/githooks-cli" installer "${EXTRA_INSTALL_ARGS[@]}" 2>&1)

# shellcheck disable=SC2181
if [ $? -ne 0 ] || echo "$OUTPUT" | grep -qi "automatic update checks"; then
    echo "! Automatic updates should have been set up already:"
    echo "$OUTPUT"
    exit 1
fi
