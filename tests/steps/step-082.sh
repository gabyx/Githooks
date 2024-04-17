#!/usr/bin/env bash
# Test:
#   Cli tool: manage global shared hook repositories

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

init_step

accept_all_trust_prompts || exit 1

if ! "$GH_TEST_BIN/githooks-cli" installer "${EXTRA_INSTALL_ARGS[@]}"; then
    echo "! Failed to execute the install script"
    exit 1
fi

mkdir -p "$GH_TEST_TMP/shared/first-shared.git/.githooks/pre-commit" &&
    mkdir -p "$GH_TEST_TMP/shared/second-shared.git/.githooks/pre-commit" &&
    mkdir -p "$GH_TEST_TMP/shared/third-shared.git/.githooks/pre-commit" &&
    echo 'echo "Hello"' >"$GH_TEST_TMP/shared/first-shared.git/.githooks/pre-commit/sample-one" &&
    echo 'echo "Hello"' >"$GH_TEST_TMP/shared/second-shared.git/.githooks/pre-commit/sample-two" &&
    echo 'echo "Hello"' >"$GH_TEST_TMP/shared/third-shared.git/.githooks/pre-commit/sample-three" &&
    (cd "$GH_TEST_TMP/shared/first-shared.git" && git init && git add . && git commit -m 'Testing') &&
    (cd "$GH_TEST_TMP/shared/second-shared.git" && git init && git add . && git commit -m 'Testing') &&
    (cd "$GH_TEST_TMP/shared/third-shared.git" && git init && git add . && git commit -m 'Testing') ||
    exit 1

mkdir -p "$GH_TEST_TMP/test082" &&
    cd "$GH_TEST_TMP/test082" &&
    git init || exit 1

function test_shared() {

    url1="file://$GH_TEST_TMP/shared/first-shared.git"
    location1=$("$GH_INSTALL_BIN_DIR/githooks-cli" shared root-from-url "$url1") || exit 1

    "$GH_INSTALL_BIN_DIR/githooks-cli" shared add --global "$url1" || exit 1
    "$GH_INSTALL_BIN_DIR/githooks-cli" shared list | grep "first-shared" | grep "pending" || exit 2
    "$GH_INSTALL_BIN_DIR/githooks-cli" shared pull || exit 3
    "$GH_INSTALL_BIN_DIR/githooks-cli" shared list | grep "first-shared" | grep "active" || exit 4
    "$GH_INSTALL_BIN_DIR/githooks-cli" shared add --global file://"$GH_TEST_TMP/shared/second-shared.git" || exit 5
    "$GH_INSTALL_BIN_DIR/githooks-cli" shared add --global file://"$GH_TEST_TMP/shared/third-shared.git" || exit 6
    "$GH_INSTALL_BIN_DIR/githooks-cli" shared list --global | grep "second-shared" | grep "pending" || exit 7
    "$GH_INSTALL_BIN_DIR/githooks-cli" shared list --all | grep "third-shared" | grep "pending" || exit 8

    (cd "$location1" &&
        git remote rm origin &&
        git remote add origin /some/other/url.git) || exit 9
    "$GH_INSTALL_BIN_DIR/githooks-cli" shared list | grep "first-shared" | grep "invalid" || exit 10
    "$GH_INSTALL_BIN_DIR/githooks-cli" shared remove --global file://"$GH_TEST_TMP/shared/first-shared.git" || exit 11
    ! "$GH_INSTALL_BIN_DIR/githooks-cli" shared list | grep "first-shared" || exit 12
    "$GH_INSTALL_BIN_DIR/githooks-cli" shared remove --global file://"$GH_TEST_TMP/shared/second-shared.git" || exit 13
    "$GH_INSTALL_BIN_DIR/githooks-cli" shared remove --global file://"$GH_TEST_TMP/shared/third-shared.git" || exit 14
    [ -z "$(git config --global --get-all githooks.shared)" ] || exit 15
}

test_shared

"$GH_INSTALL_BIN_DIR/githooks-cli" shared clear --all &&
    "$GH_INSTALL_BIN_DIR/githooks-cli" shared purge ||
    exit 16

test_shared
