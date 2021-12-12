#!/usr/bin/env bash
# Test:
#   Set up bare repos, run the install and verify the hooks get installed/uninstalled

TEST_DIR=$(cd "$(dirname "$0")" && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

if echo "$EXTRA_INSTALL_ARGS" | grep -q "use-core-hookspath"; then
    echo "Using core.hooksPath"
    exit 249
fi

acceptAllTrustPrompts || exit 1

mkdir -p "$GH_TEST_TMP/test109/p001" && mkdir -p "$GH_TEST_TMP/test109/p002" && mkdir -p "$GH_TEST_TMP/test109/p003" || exit 1

cd "$GH_TEST_TMP/test109/p001" && git init --bare || exit 1
cd "$GH_TEST_TMP/test109/p002" && git init --bare || exit 1

if grep -r 'github.com/gabyx/githooks' "$GH_TEST_TMP/test109/"; then
    echo "! Hooks were installed ahead of time"
    exit 1
fi

mkdir -p ~/.githooks/templates/hooks
git config --global init.templateDir ~/.githooks/templates

# run the install, and select installing hooks into existing repos
echo "n
y
$GH_TEST_TMP/test109
" | "$GH_TEST_BIN/cli" installer --stdin || exit 1

if ! grep -qr 'github.com/gabyx/githooks' "$GH_TEST_TMP/test109/p001/hooks" ||
    ! grep -qr 'github.com/gabyx/githooks' "$GH_TEST_TMP/test109/p002/hooks"; then
    echo "! Hooks were not installed successfully"
    exit 1
fi

# check if only server hooks are installed.
for hook in pre-push pre-receive update post-receive post-update push-to-checkout pre-auto-gc; do
    if [ ! -f "$GH_TEST_TMP/test109/p001/hooks/"$hook ]; then
        echo "! Server hooks were not installed successfully ('$hook')"
        exit 1
    fi
done
# shellcheck disable=SC2012
count=$(find "$GH_TEST_TMP/test109/p001/hooks/" -type f | wc -l)
if [ "$count" != "8" ]; then
    echo "! Expected only server hooks to be installed ($count)"
    exit 1
fi

cd "$GH_TEST_TMP/test109/p003" && git init --bare || exit 1
# check if only server hooks are installed.
for hook in pre-push pre-receive update post-receive post-update push-to-checkout pre-auto-gc; do
    if [ ! -f "$GH_TEST_TMP/test109/p003/hooks/"$hook ]; then
        echo "! Server hooks were not installed successfully ('$hook')"
        exit 1
    fi
done

echo "y
$GH_TEST_TMP/test109
" | "$GH_TEST_BIN/cli" uninstaller --stdin || exit 1

if grep -qr 'github.com/gabyx/githooks' "$GH_TEST_TMP/test109/p001/hooks" ||
    grep -qr 'github.com/gabyx/githooks' "$GH_TEST_TMP/test109/p002/hooks"; then
    echo "! Hooks were not uninstalled successfully"
    exit 1
fi
