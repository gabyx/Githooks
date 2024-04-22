#!/usr/bin/env bash
# Test:
#   Uninstall: start directory options for existing repos

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

init_step

accept_all_trust_prompts || exit 1

mkdir -p "$GH_TEST_TMP/test074/.githooks/pre-commit" &&
    echo 'echo "Hello"' >"$GH_TEST_TMP/test074/.githooks/pre-commit/testing" &&
    cd "$GH_TEST_TMP/test074" &&
    git init || exit 1

echo "y

y
y
$GH_TEST_TMP
" | "$GH_TEST_BIN/githooks-cli" installer "${EXTRA_INSTALL_ARGS[@]}" --stdin || exit 1

echo 'y

' | "$GH_TEST_BIN/githooks-cli" uninstaller --stdin || exit 2

echo 'y
/not/found
' | "$GH_TEST_BIN/githooks-cli" uninstaller --stdin

# shellcheck disable=SC2181
if [ $? -eq 0 ]; then
    echo "! Uninstall unexpectedly finished"
    exit 1
fi
