#!/usr/bin/env bash
# Test:
#   Run an install, and let it set up a new template directory (non-tilde)

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

accept_all_trust_prompts || exit 1

# delete the built-in git template folder
rm -rf "$GH_TEST_GIT_CORE/templates" || exit 1

# run the install, and let it search for the hooks dir and the chose the given one
echo "n
y
$GH_TEST_TMP/.test-020/hooks
" | "$GH_TEST_BIN/githooks-cli" installer --stdin || exit 1

mkdir -p "$GH_TEST_TMP/test20" &&
    cd "$GH_TEST_TMP/test20" &&
    git init || exit 1

if grep -q 'github.com/gabyx/githooks' "$GH_TEST_TMP/test20/.git/hooks/pre-commit"; then
    echo "! Githooks were installed into a new repo, but should have not"
    exit 1
fi

if echo "${EXTRA_INSTALL_ARGS:-}" | grep -q "centralized"; then
    if [ "$GH_TEST_TMP/.test-020/hooks" != "$(git config --global core.hooksPath)" ]; then
        echo "! Config 'core.hooksPath' does not point to the same directory."
        git config --global core.hooksPath
        exit 1
    fi
else
    git hooks install

    if [ "$GH_TEST_TMP/.test-020/hooks" != "$(git config --local core.hooksPath)" ]; then
        echo "! Config 'core.hooksPath' does not point to the same directory."
        git config --local core.hooksPath
        exit 1
    fi
fi
