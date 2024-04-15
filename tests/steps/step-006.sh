#!/usr/bin/env bash
# Test:
#   Run an install, and let it search for the template dir

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

accept_all_trust_prompts || exit 1

# move the built-in git template folder
mkdir -p "$GH_TEST_TMP/git-templates" &&
    mv "$GH_TEST_GIT_CORE/templates" "$GH_TEST_TMP/git-templates/" &&
    rm -f "$GH_TEST_TMP/git-templates/templates/hooks/"* &&
    touch "$GH_TEST_TMP/git-templates/templates/hooks/pre-commit.sample" ||
    exit 1

# run the install, and let it search for the templates
echo 'y
y
y
y
' | "$GH_TEST_BIN/githooks-cli" installer --stdin || exit 1

check_install

if ! [ -f "$GH_TEST_TMP/git-templates/templates/hooks/pre-commit" ]; then
    echo "! Expected hook is not installed"
    exit 1
fi

if ! grep 'github.com/gabyx/githooks' "$GH_TEST_TMP/git-templates/templates/hooks/pre-commit"; then
    echo "! Expected hook doesn't have the expected contents"
    exit 1
fi

mkdir -p "$GH_TEST_TMP/test6" &&
    cd "$GH_TEST_TMP/test6" &&
    git init || exit 1

if echo "${EXTRA_INSTALL_ARGS:-}" | grep -q "centralized"; then
    check_centralized_install
else
    git hooks install
    check_local_install
fi
