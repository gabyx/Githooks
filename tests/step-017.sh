#!/bin/sh
# Test:
#   Direct runner execution: execute a previously saved hook

TEST_DIR=$(cd "$(dirname "$0")" && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

acceptAllTrustPrompts || exit 1

mkdir -p "$GH_TEST_TMP/test017" &&
    cd "$GH_TEST_TMP/test017" &&
    git init || exit 1

mkdir -p .githooks/pre-commit &&
    echo "echo 'Direct execution' >> '$GH_TEST_TMP/test017.out'" >.githooks/pre-commit/test &&
    echo '#!/bin/sh' >.git/hooks/pre-commit.replaced.githook &&
    echo "echo 'Previous hook' >> '$GH_TEST_TMP/test017.out'" >>.git/hooks/pre-commit.replaced.githook &&
    chmod +x .git/hooks/pre-commit.replaced.githook &&
    "$GH_TEST_BIN/runner" "$(pwd)"/.git/hooks/pre-commit ||
    exit 1

if ! grep -q 'Direct execution' "$GH_TEST_TMP/test017.out"; then
    echo "! Direct execution didn't happen"
    exit 1
fi

if ! grep -q 'Previous hook' "$GH_TEST_TMP/test017.out"; then
    echo "! Previous hook was not executed"
    exit 1
fi
