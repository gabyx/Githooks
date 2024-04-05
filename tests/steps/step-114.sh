#!/usr/bin/env bash
# shellcheck disable=SC1091
# Test:
#   Test template area is set up properly (core.hooksPath)

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

accept_all_trust_prompts || exit 1

if echo "${EXTRA_INSTALL_ARGS:-}" | grep -q "use-core-hookspath"; then
    echo "Using core.hooksPath"
    exit 249
fi

mkdir -p "$GH_TEST_TMP/test114/.githooks/pre-commit" &&
    echo "echo 'Testing 114' > '$GH_TEST_TMP/test114.out'" >"$GH_TEST_TMP/test114/.githooks/pre-commit/test-hook" &&
    cd "$GH_TEST_TMP/test114" &&
    git init || exit 1

if grep -r 'github.com/gabyx/githooks' "$GH_TEST_TMP/test114/.git"; then
    echo "! Hooks were installed ahead of time"
    exit 2
fi

mkdir -p ~/.githooks/templates

# run the install, and select installing hooks into existing repos
echo "n
y
$GH_TEST_TMP/test114
" | "$GH_TEST_BIN/githooks-cli" installer --stdin --use-core-hookspath --template-dir ~/.githooks/templates || exit 3

# check if hooks are inside the template folder.
if ! "$GH_INSTALL_BIN_DIR/githooks-cli" list | grep -q "test-hook"; then
    echo "! Hooks were not installed successfully"
    exit 4
fi

git add . && git commit -m 'Test commit' || exit 5

if ! grep 'Testing 114' "$GH_TEST_TMP/test114.out"; then
    echo "! Expected hook did not run"
    exit 6
fi

# Reset to trigger update
if ! git -C "$GH_TEST_REPO" reset --hard v9.9.1 >/dev/null; then
    echo "! Could not reset server to trigger update."
    exit 1
fi

rm -rf ~/.githooks/templates/hooks/* # Remove to see if the correct folder gets choosen

if ! "$GH_INSTALL_BIN_DIR/githooks-cli" update --yes; then
    echo "! Failed to run the update"
    exit 1
fi

if [ ! -f ~/.githooks/templates/hooks/pre-commit ]; then
    echo "! Expected update to install wrappers into \`~/.githooks/templates\`"
    exit 1
fi
