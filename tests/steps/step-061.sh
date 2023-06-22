#!/usr/bin/env bash
# Test:
#   Cli tool: update shared hook repos

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

acceptAllTrustPrompts || exit 1

"$GH_TEST_BIN/cli" installer || exit 1

mkdir -p "$GH_TEST_TMP/shared/first-shared.git/pre-commit" &&
    cd "$GH_TEST_TMP/shared/first-shared.git" &&
    echo 'echo "Hello"' >pre-commit/sample-one &&
    git init --template= && git add . && git commit -m 'Testing' || exit 1

mkdir -p "$GH_TEST_TMP/shared/second-shared.git/pre-commit" &&
    cd "$GH_TEST_TMP/shared/second-shared.git" &&
    echo 'echo "Hello"' >pre-commit/sample-two &&
    git init --template= && git add . && git commit -m 'Testing' || exit 1

mkdir -p "$GH_TEST_TMP/shared/third-shared.git/pre-commit" &&
    cd "$GH_TEST_TMP/shared/third-shared.git" &&
    echo 'echo "Hello"' >pre-commit/sample-three &&
    git init --template= && git add . && git commit -m 'Testing' || exit 1

mkdir -p "$GH_TEST_TMP/test061/.githooks" &&
    cd "$GH_TEST_TMP/test061" &&
    echo "urls: - $GH_TEST_TMP/shared/first-shared.git" >.githooks/.shared.yaml &&
    git init ||
    exit 1

git config --local githooks.shared "file://$GH_TEST_TMP/shared/second-shared.git" || exit 1
git config --global githooks.shared "file://$GH_TEST_TMP/shared/third-shared.git" || exit 1

if ! "$GH_INSTALL_BIN_DIR/cli" list | grep "sample-one" | grep -q "'shared:repo'"; then
    echo "! Unexpected cli list output (1)"
    exit 1
fi

if "$GH_INSTALL_BIN_DIR/cli" list | grep -q "sample-two" ||
    "$GH_INSTALL_BIN_DIR/cli" list | grep -q "sample-three"; then
    echo "! Unexpected cli list output (2)"
    exit 1
fi

if ! "$GH_INSTALL_BIN_DIR/cli" list | grep -qi "pending shared hooks"; then
    echo "! Unexpected cli list output (3)"
    exit 1
fi

if ! "$GH_INSTALL_BIN_DIR/cli" shared update; then
    echo "! Failed to update the shared hook repositories"
    exit 1
fi

if ! "$GH_INSTALL_BIN_DIR/cli" list | grep "sample-one" | grep -q "'shared:repo'"; then
    echo "! Unexpected cli list output (4)"
    exit 1
fi

if ! "$GH_INSTALL_BIN_DIR/cli" list | grep "sample-two" | grep -q "'shared:local'"; then
    echo "! Unexpected cli list output (5)"
    exit 1
fi

if ! "$GH_INSTALL_BIN_DIR/cli" list | grep "sample-three" | grep -q "'shared:global'"; then
    echo "! Unexpected cli list output (6)"
    exit 1
fi
