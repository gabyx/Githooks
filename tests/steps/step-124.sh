#!/usr/bin/env bash
# Test:
#   Warning about core.hooksPath masking Githooks hook runners in the current repo

if echo "${EXTRA_INSTALL_ARGS:-}" | grep -q "use-core-hookspath"; then
    echo "Using core.hooksPath"
    exit 249
fi

if ! "$GH_TEST_BIN/cli" installer; then
    echo "! Failed to execute the install script"
    exit 1
fi

mkdir -p "$GH_TEST_TMP/test124/.githooks/pre-commit" &&
    cd "$GH_TEST_TMP/test124" &&
    echo 'echo Testing' >.githooks/pre-commit/test-pre-commit &&
    git init ||
    exit 1

cd "$GH_TEST_TMP/test124" || exit 1

if ! "$GH_TEST_BIN/cli" list | grep -q 'test-pre-commit'; then
    echo "! Expected to have the test hooks listed" >&2
    exit 2
fi

if "$GH_TEST_BIN/cli" list 2>&1 | grep -q 'hooks in this repository are not run by Githooks'; then
    echo "! Expected NOT to have a warning displayed" >&2
    exit 3
fi

git config core.hooksPath /tmp/corehooks || exit 4

if ! "$GH_TEST_BIN/cli" list | grep -q 'test-pre-commit'; then
    echo "! Expected to have the test hooks listed" >&2
    exit 5
fi

if ! "$GH_TEST_BIN/cli" list 2>&1 | grep -q 'hooks in this repository are not run by Githooks'; then
    echo "! Expected to have a warning displayed" >&2
    exit 6
fi
