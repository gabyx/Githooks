#!/usr/bin/env bash
# Test:
#   Run an install, and let it search for the template dir

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

acceptAllTrustPrompts || exit 1

if echo "$EXTRA_INSTALL_ARGS" | grep -q "use-core-hookspath"; then
    echo "Using core.hooksPath"
    exit 249
fi

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
' | "$GH_TEST_BIN/cli" installer --stdin || exit 1

if ! [ -f "$GH_TEST_TMP/git-templates/templates/hooks/pre-commit" ]; then
    # verify that a new hook file was installed
    echo "! Expected hook is not installed"
    exit 1
elif ! grep 'github.com/gabyx/githooks' "$GH_TEST_TMP/git-templates/templates/hooks/pre-commit"; then
    # verify that the new hook is ours
    echo "! Expected hook doesn't have the expected contents"
    exit 1
fi

mkdir -p "$GH_TEST_TMP/test6" &&
    cd "$GH_TEST_TMP/test6" &&
    git init || exit 1

# verify that the hooks are installed and are working
if ! grep 'github.com/gabyx/githooks' "$GH_TEST_TMP/test6/.git/hooks/pre-commit"; then
    echo "! Githooks were not installed into a new repo"
    exit 1
fi
