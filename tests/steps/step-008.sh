#!/usr/bin/env bash
# Test:
#   Run an install that preserves an existing hook in the templates directory

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

accept_all_trust_prompts || exit 1

mkdir -p ~/.githooks/mytemplates/hooks

cd ~/.githooks/mytemplates/hooks &&
    echo '#!/bin/sh' >>pre-commit &&
    echo "echo 'Previous' >> '$GH_TEST_TMP/test-008.out'" >>pre-commit &&
    chmod +x pre-commit ||
    exit 1

"$GH_TEST_BIN/githooks-cli" installer --hooks-dir ~/.githooks/mytemplates/hooks || exit 1

mkdir -p "$GH_TEST_TMP/test8/.githooks/pre-commit" &&
    cd "$GH_TEST_TMP/test8" &&
    echo "echo 'In-repo' >> '$GH_TEST_TMP/test-008.out'" >.githooks/pre-commit/test &&
    git init &&
    git add .githooks/pre-commit/test ||
    exit 1

if ! echo "${EXTRA_INSTALL_ARGS:-}" | grep -q "centralized"; then
    git commit -m 'Test'

    if grep 'Previous' "$GH_TEST_TMP/test-008.out" ||
        grep 'In-repo' "$GH_TEST_TMP/test-008.out"; then
        echo '! No hooks should have been run.'
        exit 1
    fi

    git hooks install || exit 1
fi

git commit --allow-empty -m 'Test'

if ! grep 'Previous' "$GH_TEST_TMP/test-008.out"; then
    echo '! Saved hook was not run'
    exit 1
fi

if ! grep 'In-repo' "$GH_TEST_TMP/test-008.out"; then
    echo '! Newly added hook was not run'
    exit 1
fi
