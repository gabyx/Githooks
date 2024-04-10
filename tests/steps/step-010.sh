#!/usr/bin/env bash
# Test:
#   Execute a dry-run installation

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

accept_all_trust_prompts || exit 1

mkdir -p "$GH_TEST_TMP/test10/a" &&
    cd "$GH_TEST_TMP/test10/a" &&
    git init || exit 1

echo "n
y
$GH_TEST_TMP
" | "$GH_TEST_BIN/githooks-cli" installer --stdin --dry-run --hooks-dir ~/.githooks/mytemplates || exit 1

if git config --global --get-regexp "^githooks.*" | grep -qv "deletedetectedlfshooks" ||
    [ -n "$(git config --global alias.hooks)" ]; then

    echo "Should not have set Git config variables."
    git config --global --get-regexp "^githooks.*" | grep -v "deletedetectedlfshooks"

    exit 1
fi

if [ -d ~/.githooks/mytemplates ] || [ -d ~/.githooks/release ]; then
    echo "No folders should have been created".
    ls -al ~/.githooks/mytemplates
    ls -al ~/.githooks/release
    exit 1
fi
