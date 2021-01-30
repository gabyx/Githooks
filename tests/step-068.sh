#!/bin/sh
# Test:
#   Run the cli tool in a directory that is not a Git repository

mkdir "$GH_TEST_TMP/not-a-git-repo" && cd "$GH_TEST_TMP/not-a-git-repo" || exit 1

if ! "$GH_TEST_BIN/cli" installer; then
    echo "! Failed to execute the install script"
    exit 1
fi

if "$GITHOOKS_INSTALL_BIN_DIR/cli" list; then
    echo "! Expected to fail"
    exit 1
fi

if "$GITHOOKS_INSTALL_BIN_DIR/cli" trust; then
    echo "! Expected to fail"
    exit 1
fi

if "$GITHOOKS_INSTALL_BIN_DIR/cli" disable; then
    echo "! Expected to fail"
    exit 1
fi
