#!/usr/bin/env bash
# Test:
#   Run a simple install non-interactively and verify the hooks are in place

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

init_step

accept_all_trust_prompts || exit 1

# run the default install
"$GH_TEST_BIN/githooks-cli" installer "${EXTRA_INSTALL_ARGS[@]}" --non-interactive || exit 1
check_install

if ! is_centralized_tests; then
    mkdir -p "$GH_TEST_TMP/test1" &&
        cd "$GH_TEST_TMP/test1" &&
        git init || exit 1

    check_no_local_install .

    # Install hooks
    "$GH_INSTALL_BIN_DIR/githooks-cli" install ||
        die "Could not install hooks into repo."

    check_local_install .

else
    check_centralized_install
fi
