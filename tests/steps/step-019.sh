#!/usr/bin/env bash
# Test:
#   Run an install, and set based on a custom template directory

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

init_step

accept_all_trust_prompts || exit 1

if is_centralized_tests; then
    echo "Using centralized install"
    exit 249
fi

# shellcheck disable=SC2088
mkdir -p ~/.test-019/hooks &&
    git config --global init.templateDir '~/.test-019' ||
    exit 1

"$GH_TEST_BIN/githooks-cli" installer "${EXTRA_INSTALL_ARGS[@]}" --hooks-dir-use-template-dir || exit 1

mkdir -p "$GH_TEST_TMP/test19" &&
    cd "$GH_TEST_TMP/test19" &&
    git init || exit 1

check_normal_install "$HOME/.test-019/hooks"
check_local_install_run_wrappers "."

# Reinstall and check again
"$GH_INSTALL_BIN_DIR/githooks-cli" install
check_local_install_run_wrappers "."
