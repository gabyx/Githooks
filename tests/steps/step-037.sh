#!/usr/bin/env bash
# Test:
#   Re-enabling automatic update checks

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

init_step

accept_all_trust_prompts || exit 1

git config --global githooks.updateCheckEnabled false || exit 1
echo 'y
' | "$GH_TEST_BIN/githooks-cli" installer "${EXTRA_INSTALL_ARGS[@]}" --stdin || exit 1

if [ "$(git config --global --get githooks.updateCheckEnabled)" != "true" ]; then
    echo "! Automatic update checks are not enabled"
    exit 1
fi
