#!/usr/bin/env bash
# Test:
#   Run an install, adding the intro README files into an existing repo

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

acceptAllTrustPrompts || exit 1

if echo "$EXTRA_INSTALL_ARGS" | grep -q "use-core-hookspath"; then
    echo "Using core.hooksPath"
    exit 249
fi

mkdir -p "$GH_TEST_TMP/test046/.githooks/pre-commit" &&
    echo "echo 'Testing' > '$GH_TEST_TMP/test46.out'" >"$GH_TEST_TMP/test046/.githooks/pre-commit/test" &&
    cd "$GH_TEST_TMP/test046" ||
    exit 1

git init || exit 1

echo "n
y
$GH_TEST_TMP/test046
y
" | "$GH_TEST_BIN/cli" installer --stdin || exit 1

if ! grep "github.com/gabyx/githooks" "$GH_TEST_TMP/test046/.git/hooks/pre-commit"; then
    echo "! Hooks were not installed"
    exit 1
fi

if ! grep "github.com/gabyx/githooks" "$GH_TEST_TMP/test046/.githooks/README.md"; then
    echo "! README was not installed"
    exit 1
fi
