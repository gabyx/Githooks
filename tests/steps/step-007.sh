#!/usr/bin/env bash
# Test:
#   Run an install, and let it set up a new template directory

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

accept_all_trust_prompts || exit 1

# delete the built-in git template folder
rm -rf "$GH_TEST_GIT_CORE/templates" || exit 1

# run the install, and let it search for the templates
echo 'y
' | "$GH_TEST_BIN/githooks-cli" installer --stdin || exit 1

mkdir -p "$GH_TEST_TMP/test7" &&
    cd "$GH_TEST_TMP/test7" &&
    git init || exit 1

check_install "$HOME/.githooks/templates/hooks"

if echo "${EXTRA_INSTALL_ARGS:-}" | grep -q "centralized"; then
    check_centralized_install
else
    "$GH_TEST_BIN/githooks-cli" install || exit 1
    check_local_install
fi
