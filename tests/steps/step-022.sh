#!/usr/bin/env bash
# Test:
#   Set up local repos, run the install and verify the hooks get installed (default directory)

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

accept_all_trust_prompts || exit 1

if echo "${EXTRA_INSTALL_ARGS:-}" | grep -q "centralized"; then
    echo "Using centralized install"
    exit 249
fi

mkdir -p ~/test022/p001 &&
    cd ~/test022/p001 &&
    git init || exit 1
mkdir -p ~/test022/p002 &&
    cd ~/test022/p002 &&
    git init || exit 1

# run the install, and select installing the hooks into existing repos
echo 'n
y

' | "$GH_TEST_BIN/githooks-cli" installer --stdin || exit 1

path=$(git config --global githooks.pathForUseCoreHooksPath)
[ -d "$path" ] || {
    echo "! Path '$path' does not exist."
    exit 1
}

if echo "${EXTRA_INSTALL_ARGS:-}" | grep -q "centralized"; then
    if [ "$path" != "$(git config --global core.hooksPath)" ]; then
        echo "! Config 'core.hooksPath' does not point to the same directory."
        git config --global core.hooksPath
        exit 1
    fi
else
    git hooks install

    if [ "$path" != "$(git config --local core.hooksPath)" ]; then
        echo "! Config 'core.hooksPath' does not point to the same directory."
        git config --local core.hooksPath
        exit 1
    fi
fi

rm -rf ~/test022
