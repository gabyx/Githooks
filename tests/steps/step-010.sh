#!/usr/bin/env bash
# Test:
#   Execute a dry-run installation

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

init_step

accept_all_trust_prompts || exit 1

mkdir -p "$GH_TEST_TMP/test10/a" &&
    cd "$GH_TEST_TMP/test10/a" &&
    git init || exit 1

echo "n
y
$GH_TEST_TMP
y
/tmp/test

" | "$GH_TEST_BIN/githooks-cli" installer "${EXTRA_INSTALL_ARGS[@]}" --stdin --dry-run --hooks-dir ~/.githooks/mytemplates || exit 1

# Check for all githooks config vars set (local and global)
if git config --get-regexp "^githooks.*" ||
    [ -n "$(git config --global alias.hooks)" ]; then

    echo "Should not have set Git config variables."
    git config --get-regexp "^githooks.*"

    exit 1
fi

if [ -d ~/.githooks/mytemplates ] || [ -d ~/.githooks/release ]; then
    echo "No folders should have been created".
    ls -al ~/.githooks/mytemplates
    ls -al ~/.githooks/release
    exit 1
fi
