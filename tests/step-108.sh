#!/bin/sh
# Test:
#   Custom install prefix test

TEST_DIR=$(cd "$(dirname "$0")" && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

acceptAllTrustPrompts || exit 1

TEST_PREFIX_DIR=""$GH_TEST_TMP/githooks""
GH_INSTALL_BIN_DIR="$TEST_PREFIX_DIR/.githooks/bin"

"$GH_TEST_BIN/cli" installer --prefix "$TEST_PREFIX_DIR" || exit 1

if [ ! -d "$TEST_PREFIX_DIR/.githooks" ]; then
    echo "! Expected the install directory to be in \`$TEST_PREFIX_DIR\`"
    exit 2
fi

if [ "$(git config --global githooks.installDir)" != "$TEST_PREFIX_DIR/.githooks" ]; then
    echo "! Install directory in config \`$(git config --global githooks.installDir)\` is incorrect!"
    exit 3
fi

# Set a wrong install
git config --global githooks.installDir "$TEST_PREFIX_DIR/.githooks-notexisting"

if ! "$GH_INSTALL_BIN_DIR/cli" --help 2>&1 | grep -q "Githooks installation is corrupt"; then
    echo "! Expected the installation to be corrupt"
    "$GH_INSTALL_BIN_DIR/cli" --help
    exit 4
fi

mkdir -p "$GH_TEST_TMP/test108/.githooks/pre-commit" &&
    echo 'echo "Hello"' >"$GH_TEST_TMP/test108/.githooks/pre-commit/testing" &&
    cd "$GH_TEST_TMP/test108" &&
    git init || exit 5

echo A >A.txt
git add A.txt
if ! git commit -a -m "Test" 2>&1 | grep -q "Githooks installation is corrupt"; then
    echo "! Expected the installation to be corrupt"
    exit 6
fi
