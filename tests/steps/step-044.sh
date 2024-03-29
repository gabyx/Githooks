#!/usr/bin/env bash
# Test:
#   Run an install including the intro README files for one repo

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

acceptAllTrustPrompts || exit 1

if echo "${EXTRA_INSTALL_ARGS:-}" | grep -q "use-core-hookspath"; then
    echo "Using core.hooksPath"
    exit 249
fi

mkdir -p "$GH_TEST_TMP/test044/001" &&
    cd "$GH_TEST_TMP/test044/001" &&
    git init || exit 1
mkdir -p "$GH_TEST_TMP/test044/002" &&
    cd "$GH_TEST_TMP/test044/002" &&
    git init || exit 1

cd "$GH_TEST_TMP/test044" || exit 1

echo "n
y
$GH_TEST_TMP/test044
n
y
" | "$GH_TEST_BIN/cli" installer --stdin || exit 1

find "$GH_TEST_TMP/test044"

if ! grep "github.com/gabyx/githooks" "$GH_TEST_TMP/test044/001/.git/hooks/pre-commit"; then
    echo "! Hooks were not installed into 001"
    exit 1
fi

if grep "github.com/gabyx/githooks" "$GH_TEST_TMP/test044/001/.githooks/README.md"; then
    echo "! README was unexpectedly installed into 001"
    exit 1
fi

if ! grep "github.com/gabyx/githooks" "$GH_TEST_TMP/test044/002/.git/hooks/pre-commit"; then
    echo "! Hooks were not installed into 002"
    exit 1
fi

if ! grep "github.com/gabyx/githooks" "$GH_TEST_TMP/test044/002/.githooks/README.md"; then
    echo "! README was not installed into 002"
    exit 1
fi
