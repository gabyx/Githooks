#!/bin/sh
# Test:
#   Warning about core.hooksPath not being used

if ! echo "${EXTRA_INSTALL_ARGS:-}" | grep -q "use-core-hookspath"; then
    echo "Test needs core.hooksPath to be configured"
    exit 249
fi

mkdir -p "$GH_TEST_TMP/test125-core-hookspath" || exit 1
git config --global core.hooksPath "$GH_TEST_TMP/test125-core-hookspath"

if ! "$GH_TEST_BIN/githooks-cli" installer --use-core-hookspath; then
    echo "! Failed to execute the install script"
    exit 1
fi

mkdir -p "$GH_TEST_TMP/test125/.githooks/pre-commit" &&
    cd "$GH_TEST_TMP/test125" &&
    echo 'echo Testing' >.githooks/pre-commit/test-pre-commit &&
    git init ||
    exit 1

cd "$GH_TEST_TMP/test125" || exit 1

if ! "$GH_TEST_BIN/githooks-cli" list | grep -q 'test-pre-commit'; then
    echo "! Expected to have the test hooks listed" >&2
    exit 2
fi

if "$GH_TEST_BIN/githooks-cli" list 2>&1 | grep -q 'hooks in this repository are not run by Githooks'; then
    echo "! Expected NOT to have a warning displayed" >&2
    exit 3
fi

git config --global --unset core.hooksPath || exit 4

if ! "$GH_TEST_BIN/githooks-cli" list | grep -q 'test-pre-commit'; then
    echo "! Expected to have the test hooks listed" >&2
    exit 5
fi

if ! "$GH_TEST_BIN/githooks-cli" list 2>&1 | grep -q 'hooks in this repository are not run by Githooks'; then
    echo "! Expected to have a warning displayed" >&2
    exit 6
fi
