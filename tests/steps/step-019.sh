#!/usr/bin/env bash
# Test:
#   Run an install, and set based on a custom template directory

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

accept_all_trust_prompts || exit 1

# shellcheck disable=SC2088
mkdir -p ~/.test-019/hooks &&
    git config --global init.templateDir '~/.test-019' ||
    exit 1

"$GH_TEST_BIN/githooks-cli" installer || exit 1

mkdir -p "$GH_TEST_TMP/test19" &&
    cd "$GH_TEST_TMP/test19" &&
    git init || exit 1

# verify that the hooks are installed and are working
if ! grep -q 'github.com/gabyx/githooks' "$GH_TEST_TMP/test19/.git/hooks/pre-commit"; then
    echo "! Githooks were not installed into a new repo"
    exit 1
fi

if echo "${EXTRA_INSTALL_ARGS:-}" | grep -q "centralized"; then
    check_global_install_correct "$HOME/.test-019/hooks"
else
    git hooks install
    check_local_install_correct "." "$HOME/.test-019/hooks"
fi
