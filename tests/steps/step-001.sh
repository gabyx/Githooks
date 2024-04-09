#!/usr/bin/env bash
# Test:
#   Run a simple install non-interactively and verify the hooks are in place

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

accept_all_trust_prompts || exit 1

# run the default install
"$GH_TEST_BIN/githooks-cli" installer --non-interactive || exit 1

# Verify that hooks are installed.
path=$(git config --global githooks.pathForUseCoreHooksPath)
if ! grep -q 'https://github.com/gabyx/githooks' "$path/pre-commit"; then
    echo "Did not find hooks"
    exit 1
fi

if ! echo "${EXTRA_INSTALL_ARGS:-}" | grep -q "centralized"; then
    mkdir -p "$GH_TEST_TMP/test1" &&
        cd "$GH_TEST_TMP/test1" &&
        git init || exit 1

    # Install hooks
    git hooks install || {
        echo "Could not install hooks into repo."
        exit 1
    }

    if [ "$path" != "$(git config --local core.hooksPath)" ]; then
        echo "Config 'core.hooksPath' does not point to the same directory."
        exit 1
    fi

else
    if [ "$path" != "$(git config --global core.hooksPath)" ]; then
        echo "Config 'core.hooksPath' does not point to the same directory."
        exit 1
    fi
fi
