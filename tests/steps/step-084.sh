#!/usr/bin/env bash
# Test:
#   Cli tool: shared hook repository management failures

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

accept_all_trust_prompts || exit 1

if ! "$GH_TEST_BIN/githooks-cli" installer "${EXTRA_INSTALL_ARGS[@]}"; then
    echo "! Failed to execute the install script"
    exit 1
fi

mkdir -p "$GH_TEST_TMP/test084" &&
    cd "$GH_TEST_TMP/test084" &&
    git init || exit 1

"$GH_INSTALL_BIN_DIR/githooks-cli" unknown && exit 2
"$GH_INSTALL_BIN_DIR/githooks-cli" shared add && exit 4
"$GH_INSTALL_BIN_DIR/githooks-cli" shared remove && exit 5
"$GH_INSTALL_BIN_DIR/githooks-cli" shared add --shared "$GH_TEST_TMP/some/repo.git" && exit 6
"$GH_INSTALL_BIN_DIR/githooks-cli" shared remove --shared "$GH_TEST_TMP/some/repo.git" 2>&1 | grep -q "does not exist" || exit
"$GH_INSTALL_BIN_DIR/githooks-cli" shared clear unknown && exit 9
"$GH_INSTALL_BIN_DIR/githooks-cli" shared list unknown && exit 10

exit 0
