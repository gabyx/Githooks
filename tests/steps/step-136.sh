#!/usr/bin/env bash
# Test:
#   Run manual install and check hooks are not installed
set -u

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091

. "$TEST_DIR/general.sh"

if echo "${EXTRA_INSTALL_ARGS:-}" | grep -q "centralized"; then
    echo "Using centralized install"
    exit 249
fi

"$GH_TEST_BIN/githooks-cli" installer --use-manual || exit 1

mkdir -p "$GH_TEST_TMP/test136" &&
    cd "$GH_TEST_TMP/test136" &&
    git init

if [ -n "$(git config core.hooksPath)" ]; then
    echo "Git core.hooskPath is set but should not."
    exit 1
fi

if [ -n "$(git config init.templateDir)" ]; then
    echo "Git init.templateDir is set but should not."
    exit 1
fi

if grep -Rq 'github.com/gabyx/githooks' .git/hooks; then
    echo "Hooks should not have been installed."
    exit 1
fi

# Reinstall and check that it fails (uninstall first)
out=$("$GH_TEST_BIN/githooks-cli" installer --centralized 2>&1)

# shellcheck disable=SC2181
if [ $? -eq 0 ] ||
    ! echo "$out" | grep -q -i "You seem to have already installed Githooks in mode"; then
    echo -e "Install should have failed:\n$out"
    exit 1
fi

git hooks install || exit 1

if ! grep -Rq 'github.com/gabyx/githooks' .git/hooks; then
    echo "Hooks should have been installed."
    exit 1
fi

if [ -n "$(git config core.hooksPath)" ]; then
    echo "Git core.hooskPath is set but should not."
    exit 1
fi

if [ -n "$(git config init.templateDir)" ]; then
    echo "Git init.templateDir is set but should not."
    exit 1
fi

# Uninstall and reinstall normally.
"$GH_TEST_BIN/githooks-cli" uninstaller || exit 1
"$GH_TEST_BIN/githooks-cli" installer || exit 1
