#!/usr/bin/env bash
# shellcheck disable=SC1091
# Test:
#   Disable, enable and accept a shared hook (no 'githooks' directory)

TEST_DIR=$(cd "$(dirname "$0")" && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

acceptAllTrustPrompts || exit 1

git config --global githooks.testingTreatFileProtocolAsRemote "true"

if ! "$GH_TEST_BIN/cli" installer; then
    echo "! Failed to execute the install script"
    exit 1
fi

mkdir -p "$GH_TEST_TMP/test115.shared/shared-repo.git/pre-commit" &&
    cd "$GH_TEST_TMP/test115.shared/shared-repo.git" &&
    git init &&
    echo "echo 'Shared invoked' > '$GH_TEST_TMP/test115.out'" >pre-commit/test-shared &&
    echo "mygagahooks" >.namespace &&
    git add pre-commit .namespace &&
    git commit -m 'Initial commit' ||
    exit 2

mkdir -p "$GH_TEST_TMP/test115.repo" &&
    cd "$GH_TEST_TMP/test115.repo" &&
    git init || exit 3

"$GH_INSTALL_BIN_DIR/cli" shared add --shared file://"$GH_TEST_TMP/test115.shared/shared-repo.git" &&
    "$GH_INSTALL_BIN_DIR/cli" shared list | grep "shared-repo" | grep "pending" &&
    "$GH_INSTALL_BIN_DIR/cli" shared pull || exit 4

if ! "$GH_INSTALL_BIN_DIR/cli" list | grep 'test-shared' | grep 'shared:repo' | grep "'active'" | grep "'untrusted'"; then
    "$GH_INSTALL_BIN_DIR/cli" list
    exit 5
fi

"$GH_INSTALL_BIN_DIR/cli" ignore add --pattern 'ns:mygagahooks/**/test-shared' ||
    exit 6

if ! "$GH_INSTALL_BIN_DIR/cli" list | grep 'test-shared' |
    grep "'shared:repo'" | grep "'ignored'" | grep -q "'untrusted'"; then
    echo "! Failed to ignore shared hook"
    exit 7
fi

"$GH_INSTALL_BIN_DIR/cli" ignore add --pattern '!ns:*/**/test-shared' ||
    exit 8

if ! "$GH_INSTALL_BIN_DIR/cli" list | grep 'test-shared' |
    grep "'shared:repo'" | grep "'active'" | grep -q "'untrusted'"; then
    echo "! Failed to activate shared hook"
    exit 7
fi

"$GH_INSTALL_BIN_DIR/cli" trust hooks --pattern 'ns:my*hooks/**/test-shared' ||
    exit 10

if ! "$GH_INSTALL_BIN_DIR/cli" list | grep 'test-shared' |
    grep "'shared:repo'" | grep "'active'" | grep -q "'trusted'"; then
    echo "! Failed to trust shared hook"
    exit 7
fi
