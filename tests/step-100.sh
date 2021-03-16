#!/bin/sh
# Test:
#   Set up local repos, run the install and skip installing hooks into existing directories

TEST_DIR=$(cd "$(dirname "$0")" && pwd)
# shellcheck disable=SC1090
. "$TEST_DIR/general.sh"

acceptAllTrustPrompts || exit 1

if echo "$EXTRA_INSTALL_ARGS" | grep -q "use-core-hookspath"; then
    echo "Using core.hooksPath"
    exit 249
fi

mkdir -p ~/test100/p001 ~/test100/p002 &&
    cd ~/test100/p001 &&
    git init &&
    cd ~/test100/p002 &&
    git init || exit 1

if grep -r 'github.com/gabyx/githooks' ~/test100/; then
    echo "! Hooks were installed ahead of time"
    exit 1
fi

# run the install, and skip installing the hooks into existing repos
echo 'n
y

' | "$GH_TEST_BIN/cli" installer --stdin --skip-install-into-existing || exit 1

if grep -r 'github.com/gabyx/githooks' ~/test100/; then
    echo "! Hooks were installed but shouldn't have"
    exit 1
fi

# run the install, and let it install into existing repos
echo 'n
y

' | "$GH_TEST_BIN/cli" installer --stdin

if ! grep -r 'github.com/gabyx/githooks' ~/test100/p001/.git/hooks; then
    echo "! Hooks were not installed successfully"
    exit 1
fi

if ! grep -r 'github.com/gabyx/githooks' ~/test100/p002/.git/hooks; then
    echo "! Hooks were not installed successfully"
    exit 1
fi

rm -rf ~/test100
