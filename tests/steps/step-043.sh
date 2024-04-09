#!/usr/bin/env bash
# Test:
#   Run an install including the intro README files

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

accept_all_trust_prompts || exit 1

if echo "${EXTRA_INSTALL_ARGS:-}" | grep -q "centralized"; then
    echo "Using centralized install"
    exit 249
fi

mkdir -p "$GH_TEST_TMP/test043/001" &&
    cd "$GH_TEST_TMP/test043/001" &&
    git init || exit 1
mkdir -p "$GH_TEST_TMP/test043/002" &&
    cd "$GH_TEST_TMP/test043/002" &&
    git init || exit 1

cd "$GH_TEST_TMP/test043" || exit 1

echo "n
y
$GH_TEST_TMP/test043
a
" | "$GH_TEST_BIN/githooks-cli" installer --stdin || exit 1

if ! grep "github.com/gabyx/githooks" "$GH_TEST_TMP/test043/001/.git/hooks/pre-commit"; then
    echo "! Hooks were not installed into 001"
    exit 1
fi

if ! grep "github.com/gabyx/githooks" "$GH_TEST_TMP/test043/001/.githooks/README.md"; then
    echo "! README was not installed into 001"
    exit 1
fi

if ! grep "github.com/gabyx/githooks" "$GH_TEST_TMP/test043/002/.git/hooks/pre-commit"; then
    echo "! Hooks were not installed into 002"
    exit 1
fi

if ! grep "github.com/gabyx/githooks" "$GH_TEST_TMP/test043/002/.githooks/README.md"; then
    echo "! README was not installed into 002"
    exit 1
fi
