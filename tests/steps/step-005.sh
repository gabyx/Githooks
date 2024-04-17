#!/usr/bin/env bash
# Test:
#   Run an install with shared hooks set up, and verify those trigger properly

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

init_step

accept_all_trust_prompts || exit 1

mkdir -p "$GH_TEST_TMP/shared/hooks-005.git/pre-commit" &&
    echo "echo 'From shared hook' > '$GH_TEST_TMP/test-005.out'" \
        >"$GH_TEST_TMP/shared/hooks-005.git/pre-commit/say-hello" || exit 1

cd "$GH_TEST_TMP/shared/hooks-005.git" &&
    git init &&
    git add . &&
    git commit -m 'Initial commit'

# run the install, and set up shared repos
if echo "${EXTRA_INSTALL_ARGS:-}" | grep -q "centralized"; then
    echo "y

n
y
$GH_TEST_TMP/shared/hooks-005.git
" | "$GH_TEST_BIN/githooks-cli" installer "${EXTRA_INSTALL_ARGS[@]}" --stdin || exit 1

else
    echo "y

n
n
y
$GH_TEST_TMP/shared/hooks-005.git
" | "$GH_TEST_BIN/githooks-cli" installer "${EXTRA_INSTALL_ARGS[@]}" --stdin || exit 1
fi

mkdir -p "$GH_TEST_TMP/test5" &&
    cd "$GH_TEST_TMP/test5" &&
    git init &&
    install_hooks_if_not_centralized || exit 1

# verify that the hooks are installed and are working
git commit -m 'Test'

"$GH_TEST_BIN/githooks-cli" list

if ! grep 'From shared hook' "$GH_TEST_TMP/test-005.out"; then
    echo "! The shared hooks don't seem to be working"
    exit 1
fi
