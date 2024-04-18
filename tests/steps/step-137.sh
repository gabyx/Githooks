#!/usr/bin/env bash
# Test:
#   Run install.sh script from the main branch.
set -u

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091

. "$TEST_DIR/general.sh"

init_step

if [ -n "${GH_COVERAGE_DIR:-}" ]; then
    echo "Test cannot run for coverage."
    exit 249
fi

ref="$GH_COMMIT_SHA"

# Install with current script.
# the version on the `main` branch.
curl -sL "https://raw.githubusercontent.com/gabyx/Githooks/$ref/scripts/install.sh" | bash -s -- -- || {
    echo "Could not download install.sh from '$ref'. Did you commit that?"
    exit 1
}

"$GH_INSTALL_BIN_DIR/githooks-cli" uninstaller || exit 1

# mkdir -p "$GH_TEST_TMP/test137" &&
#     cd "$GH_TEST_TMP/test137" &&
#     git init &&
#     install_hooks_if_not_centralized || exit 1
#
# if [ -z "$(git config core.hooksPath)" ]; then
#     echo "Git core.hooskPath is not set but should."
#     exit 1
# fi
#
# if [ -n "$(git config init.templateDir)" ]; then
#     echo "Git init.templateDir is set but should not."
#     exit 1
# fi
#
# if grep -Rq 'github.com/gabyx/githooks' .git/hooks; then
#     echo "Hooks should not have been installed."
#     exit 1
# fi
#
# "$GH_INSTALL_BIN_DIR/githooks-cli" uninstaller || exit 1
