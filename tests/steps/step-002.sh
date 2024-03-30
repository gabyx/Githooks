#!/usr/bin/env bash
# Test:
#   Run a simple install and verify a hook triggers properly

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

# run the default install
"$GH_TEST_BIN/cli" installer || exit 1

mkdir -p "$GH_TEST_TMP/test2" &&
    cd "$GH_TEST_TMP/test2" &&
    git init || exit 1

# add a pre-commit hook, execute and verify that it worked
mkdir -p .githooks/pre-commit &&
    echo "echo 'From githooks' > '$GH_TEST_TMP/hooktest'" >.githooks/pre-commit/test ||
    exit 2

git commit --allow-empty -m 'Test'

if grep -q 'From githooks' "$GH_TEST_TMP/hooktest" 2>/dev/null; then
    echo "Expected hook to not run"
    exit 3
fi

accept_all_trust_prompts || exit 4

git commit --allow-empty -m 'Test' || exit 5
SHA_BEFORE=$(git rev-parse HEAD 2>/dev/null)

if ! grep -q 'From githooks' "$GH_TEST_TMP/hooktest"; then
    echo "Expected hook to run"
    exit 6
fi

# Add a post-checkout to check if parameters are passed through properly
git commit --allow-empty -m 'Test' || exit 7
SHA_AFTER=$(git rev-parse HEAD 2>/dev/null)

mkdir -p .githooks/post-checkout &&
    echo "echo \"\$1\" > '$GH_TEST_TMP/hooktest-2'" >.githooks/post-checkout/test &&
    echo "echo \"\$2\" >> '$GH_TEST_TMP/hooktest-2'" >>.githooks/post-checkout/test &&
    echo "echo \"\$3\">> '$GH_TEST_TMP/hooktest-2'" >>.githooks/post-checkout/test ||
    exit 8

if ! git checkout "$SHA_BEFORE"; then
    echo "Expected checkout to work. Passing arguments fail?"
    exit 9
fi

if ! grep -q "$SHA_AFTER" "$GH_TEST_TMP/hooktest-2" ||
    ! grep -q "$SHA_BEFORE" "$GH_TEST_TMP/hooktest-2" ||
    [ "$(wc -l <"$GH_TEST_TMP/hooktest-2")" != 3 ]; then
    echo "$SHA_BEFORE, $SHA_AFTER"
    echo "Expected 3 arguments to pass to hook script. Output:"
    cat "$GH_TEST_TMP/hooktest-2"
    exit 10
fi
