#!/usr/bin/env bash
# Test:
#   Set up local repos, run the install and verify the hooks get installed

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

accept_all_trust_prompts || exit 1

if echo "${EXTRA_INSTALL_ARGS:-}" | grep -q "centralized"; then
    echo "Using centralized install"
    exit 249
fi

mkdir -p "$GH_TEST_TMP/test4/p001" && mkdir -p "$GH_TEST_TMP/test4/p002" || exit 1

cd "$GH_TEST_TMP/test4/p001" &&
    git init || exit 1
cd "$GH_TEST_TMP/test4/p002" &&
    git init || exit 1

if grep -r 'github.com/gabyx/githooks' "$GH_TEST_TMP/test4/"; then
    echo "! Hooks were installed ahead of time"
    exit 1
fi

# run the install, and select installing the hooks into existing repos
echo "n
y
$GH_TEST_TMP/test4
" | "$GH_TEST_BIN/githooks-cli" installer --stdin || exit 1

if grep -r 'github.com/gabyx/githooks' "$GH_TEST_TMP/test4/p001/.git/hooks" ||
    grep -r 'github.com/gabyx/githooks' "$GH_TEST_TMP/test4/p002/.git/hooks"; then
    echo "Hooks were installed which should not happen"
    exit 1
fi

path=$(git config --global githooks.pathForUseCoreHooksPath)
if [ "$path" != "$(cd "$GH_TEST_TMP/test4/p001/.git/hooks" && git config --local core.hooksPath)" ] ||
    [ "$path" != "$(cd "$GH_TEST_TMP/test4/p002/.git/hooks" && git config --local core.hooksPath)" ]; then
    echo "Config 'core.hooksPath' does not point to the same directory."
    exit 1
fi
