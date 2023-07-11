#!/usr/bin/env bash
# Test:
#   Run the install and verify only server hooks get installed/uninstalled

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

if echo "${EXTRA_INSTALL_ARGS:-}" | grep -q "use-core-hookspath"; then
    echo "Using core.hooksPath"
    exit 249
fi

if command -v git-lfs; then
    # For the next tests we need git-lfs missing
    # otherwise Git LFS hooks will be reinstalled.
    # on a server you mostly! do not have git-lfs installed
    echo "Using git-lfs"
    exit 249
fi

acceptAllTrustPrompts || exit 1

mkdir -p ~/.githooks/templates/hooks
git config --global init.templateDir ~/.githooks/templates
templateDir=$(git config --global init.templateDir)

mkdir -p "$GH_TEST_TMP/test130" &&
    cd "$GH_TEST_TMP/test130" &&
    git init --bare || exit 1

# run the install, and select installing only server hooks into existing repos
echo "n
y
$GH_TEST_TMP/test130
" | "$GH_TEST_BIN/cli" installer --stdin --maintained-hooks "server" || exit 1

# check if only server hooks are inside the template folder.
for hook in pre-push pre-receive update post-receive post-update push-to-checkout pre-auto-gc; do
    if ! [ -f "$templateDir/hooks/$hook" ]; then
        echo "! Server hooks were not installed successfully"
        exit 1
    fi
done
# shellcheck disable=SC2012
count="$(find "$templateDir/hooks/" -type f | wc -l)"
if [ "$count" != "8" ]; then
    echo "! Expected only server hooks to be installed ($count)"
    exit 1
fi
