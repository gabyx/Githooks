#!/usr/bin/env bash
# Test:
#   PR #135: Bugfix: Test that init.templateDir is not set when using core.hooksPath.

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

init_step

accept_all_trust_prompts || exit 1

if [ "$(id -u)" != "0" ] || ! echo "${EXTRA_INSTALL_ARGS:-}" | grep -q "centralized"; then
    echo "! Test needs root access and --centralized."
    exit 249
fi

rm -rf "$GH_TEST_GIT_CORE/templates/hooks" || exit 1

"$GH_TEST_BIN/githooks-cli" installer "${EXTRA_INSTALL_ARGS[@]}" --non-interactive || exit 1

if [ -n "$(git config init.templateDir)" ]; then
    echo "! Expected to have init.templateDir not set!" >&2
    exit 1
fi
