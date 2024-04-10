#!/usr/bin/env bash
# Test:
#   Run an install, and let it set up a new template directory

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

accept_all_trust_prompts || exit 1

# delete the built-in git template folder
rm -rf "$GH_TEST_GIT_CORE/templates" || exit 1

# run the install, and let it search for the templates
echo 'n
y
' | "$GH_TEST_BIN/githooks-cli" installer --stdin || exit 1

mkdir -p "$GH_TEST_TMP/test7" &&
    cd "$GH_TEST_TMP/test7" &&
    git init || exit 1

path=$(git config --global githooks.pathForUseCoreHooksPath)

[ -d "$path" ] || {
    echo "! Path '$path' does not exist."
    exit 1
}

if [ "$path" != "$HOME/.githooks/templates/hooks" ]; then
    echo "Install into wrong directory."
    exit 1
fi

if echo "${EXTRA_INSTALL_ARGS:-}" | grep -q "centralized"; then
    if [ "$path" != "$(git config --global core.hooksPath)" ]; then
        echo "Config 'core.hooksPath' does not point to the same directory."
        exit 1
    fi
else
    git hooks install

    if [ "$path" != "$(git config --local core.hooksPath)" ]; then
        echo "Config 'core.hooksPath' does not point to the same directory."
        exit 1
    fi
fi
