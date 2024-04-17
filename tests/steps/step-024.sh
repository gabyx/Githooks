#!/usr/bin/env bash
# Test:
#   Run an install that unsets shared repositories

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

accept_all_trust_prompts || exit 1

# change it and expect it to reset it
git config --global githooks.shared "$GH_TEST_TMP/shared/some-previous-example"

# run the install, and set up shared repos
if echo "${EXTRA_INSTALL_ARGS:-}" | grep -q "centralized"; then
    echo 'y

n
y

' | "$GH_TEST_BIN/githooks-cli" installer "${EXTRA_INSTALL_ARGS[@]}" --stdin || exit 1

else
    echo 'y

n
n
y

' | "$GH_TEST_BIN/githooks-cli" installer "${EXTRA_INSTALL_ARGS[@]}" --stdin || exit 1

fi

SHARED_REPOS=$(git config --global --get githooks.shared)

if [ -n "$SHARED_REPOS" ]; then
    echo "! The shared hook repos are still set to: $SHARED_REPOS"
    exit 1
fi
