#!/usr/bin/env bash
# Test:
#   Run an install, and let it set up a new template directory

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

# run the install, and let it search for the templates
echo 'n
y
' | "$GH_TEST_BIN/cli" installer --stdin || exit 1

mkdir -p "$GH_TEST_TMP/test7" &&
    cd "$GH_TEST_TMP/test7" &&
    git init || exit 1

# verify that the hooks are installed and are working
if ! grep 'github.com/gabyx/githooks' "$GH_TEST_TMP/test7/.git/hooks/pre-commit"; then
    echo "! Githooks were not installed into a new repo"
    exit 1
fi
