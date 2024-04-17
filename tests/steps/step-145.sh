#!/usr/bin/env bash
# Test:
#   Test Git config not using absolute paths.
# shellcheck disable=SC1091

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

init_step

accept_all_trust_prompts || exit 1

if ! "$GH_TEST_BIN/githooks-cli" installer "${EXTRA_INSTALL_ARGS[@]}" --git-config-no-abs-path; then
    echo "! Failed to execute the install script"
    exit 1
fi

if [ "$(git config --global githooks.runner)" != "githooks-runner" ]; then
    echo "Not correct Git config value."
    git config --global githooks.runner
    exit 1
fi

if [ "$(git config --global githooks.dialog)" != "githooks-dialog" ]; then
    echo "Not correct Git config value."
    git config --global githooks.dialog
    exit 1
fi

if ! git config --global alias.hooks | grep -qE '!"?githooks-cli'; then
    echo "Not correct Git config value."
    git config --global alias.hooks
    exit 1
fi
