#!/bin/sh
# Test:
#   Cli tool: manage ignore files

TEST_DIR=$(cd "$(dirname "$0")" && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

acceptAllTrustPrompts || exit 1

if ! "$GH_TEST_BIN/cli" installer; then
    echo "! Failed to execute the install script"
    exit 1
fi

mkdir -p "$GH_TEST_TMP/test085" &&
    cd "$GH_TEST_TMP/test085" &&
    git init || exit 1

"$GH_INSTALL_BIN_DIR/cli" ignore add --repository --pattern "pre-commit/test-root" &&
    grep -q 'pre-commit/test-root' ".githooks/.ignore.yaml" || exit 6

"$GH_INSTALL_BIN_DIR/cli" ignore add --repository --path "pre-commit/test-second" &&
    grep -q "test-root" ".githooks/.ignore.yaml" &&
    grep -q "test-second" ".githooks/.ignore.yaml" || exit 7

"$GH_INSTALL_BIN_DIR/cli" ignore add --repository --hook-name "pre-commit" --path "test-pc" &&
    grep -q "test-pc" ".githooks/pre-commit/.ignore.yaml" || exit 7

mkdir -p ".githooks/post-commit/.ignore.yaml" &&
    ! "$GH_INSTALL_BIN_DIR/cli" ignore add --repository --hook-name "post-commit" --pattern "test-fail" &&
    [ ! -f ".githooks/post-commit/.ignore.yaml" ] || exit 8
