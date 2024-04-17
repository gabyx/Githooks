#!/usr/bin/env bash
# Test:
#   Set up local repos, run the install and verify the hooks get installed

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

init_step

accept_all_trust_prompts || exit 1

if is_centralized_tests; then
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
echo "y

n
y
$GH_TEST_TMP/test4
" | "$GH_TEST_BIN/githooks-cli" installer "${EXTRA_INSTALL_ARGS[@]}" --stdin || exit 1

check_install
check_local_install "$GH_TEST_TMP/test4/p001/.git/hooks"
check_local_install "$GH_TEST_TMP/test4/p002/.git/hooks"
