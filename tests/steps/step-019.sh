#!/usr/bin/env bash
# Test:
#   Run an install, and set based on a custom template directory

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

accept_all_trust_prompts || exit 1

if echo "${EXTRA_INSTALL_ARGS:-}" | grep -q "use-core-hookspath"; then
    echo "Using core.hooksPath"
    exit 249
fi

# delete the built-in git template folder
rm -rf "$GH_TEST_GIT_CORE/templates" || exit 1

# shellcheck disable=SC2088
mkdir -p ~/.test-019/hooks &&
    git config --global init.templateDir '~/.test-019' ||
    exit 1

"$GH_TEST_BIN/cli" installer || exit 1

mkdir -p "$GH_TEST_TMP/test19" &&
    cd "$GH_TEST_TMP/test19" &&
    git init || exit 1

# verify that the hooks are installed and are working
if ! grep 'github.com/gabyx/githooks' "$GH_TEST_TMP/test19/.git/hooks/pre-commit"; then
    echo "! Githooks were not installed into a new repo"
    exit 1
fi
