#!/usr/bin/env bash
# Test:
#   Git worktrees: run hooks

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

# shellcheck disable=SC2086
mkdir -p "$GH_TEST_TMP/test098/.git/hooks" &&
    cd "$GH_TEST_TMP/test098" &&
    git init &&
    "$GH_TEST_BIN/cli" installer &&
    git config githooks.autoUpdateEnabled false ||
    exit 1

if ! echo "$EXTRA_INSTALL_ARGS" | grep -q "use-core-hookspath"; then
    # When not using core.hooksPath we install into the current repository.
    if ! "$GH_TEST_BIN/cli" install --non-interactive; then
        echo "! Install into current repo failed"
        exit 1
    fi
fi

if ! git worktree list >/dev/null 2>/dev/null; then
    echo "Git worktree support is missing"
    exit 249
fi

# shellcheck disable=SC2016
mkdir -p .githooks/pre-commit &&
    echo "echo \"'\$(pwd)'\" > '$GH_TEST_TMP/test098.out'" >.githooks/pre-commit/test &&
    git add .githooks ||
    exit 1
expectedPwd=$(cd "$GH_TEST_TMP/test098" && pwd)

echo "test" >testing.txt &&
    git add testing.txt ||
    exit 1

ACCEPT_CHANGES=A git commit -m 'testing hooks' || exit 1

if ! grep -q "'$expectedPwd'" "$GH_TEST_TMP/test098.out"; then
    echo "! Unexpected target content != '$expectedPwd'"
    cat "$GH_TEST_TMP/test098.out"
    exit 1
fi

# Add another workspace
git worktree add -b example-a "$GH_TEST_TMP/test098-A" HEAD || exit 2
expectedPwd=$(cd "$GH_TEST_TMP/test098-A" && pwd)

cd "$GH_TEST_TMP/test098-A" &&
    echo "test: A" >testing.txt &&
    git add testing.txt ||
    exit 3

ACCEPT_CHANGES=A git commit -m 'testing hooks (from A)' || exit 3

if ! grep -q "'$expectedPwd'" "$GH_TEST_TMP/test098.out"; then
    echo "! Unexpected target content != '$expectedPwd'"
    cat "$GH_TEST_TMP/test098.out"
    exit 3
fi

# Add another workspace
git worktree add -b example-b "$GH_TEST_TMP/test098-B" HEAD || exit 2
expectedPwd=$(cd "$GH_TEST_TMP/test098-B" && pwd)

cd "$GH_TEST_TMP/test098-B" &&
    echo "test: B" >testing.txt &&
    git add testing.txt ||
    exit 4

ACCEPT_CHANGES=A git commit -m 'testing hooks (from B)' || exit 4

if ! grep -q "'$expectedPwd'" "$GH_TEST_TMP/test098.out"; then
    echo "! Unexpected target content != '$expectedPwd'"
    cat "$GH_TEST_TMP/test098.out"
    exit 4
fi
