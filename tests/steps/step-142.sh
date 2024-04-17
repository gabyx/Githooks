#!/usr/bin/env bash
# Test:
#   Run an normal install and check for wrong install usage

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

init_step

if echo "${EXTRA_INSTALL_ARGS:-}" | grep -q "centralized"; then
    echo "Not using centralized install"
    exit 249
fi

accept_all_trust_prompts || exit 1

"$GH_TEST_BIN/githooks-cli" installer "${EXTRA_INSTALL_ARGS[@]}" --prefix ~ || exit 1
check_install

mkdir -p "$GH_TEST_TMP/test8/.githooks/pre-commit" &&
    cd "$GH_TEST_TMP/test8" &&
    echo "echo 'In-repo' >> '$GH_TEST_TMP/test-009.out'" >.githooks/pre-commit/test &&
    git init &>/dev/null &&
    git add .githooks/pre-commit/test ||
    exit 1

check_normal_install

echo "- Install local"
"$GH_INSTALL_BIN_DIR/githooks-cli" install || exit 1

echo "- Check that pre-commit runs."
git commit -m 'Test' &>/dev/null || exit 1
if ! grep 'In-repo' "$GH_TEST_TMP/test-009.out"; then
    echo '! Hooks should have been run.'
    exit 1
fi

echo "- Installer 1"
# Install again but with different install
# Wrap cmd on new line to prohibit regex-replacement for add arguments.
OUT=$("$GH_TEST_BIN/githooks-cli" \
    installer --centralized 2>&1)
# shellcheck disable=SC2181
if [ "$?" -eq "0" ] || ! echo "$OUT" | grep -qiE "You seem to have already installed Githooks in mode"; then
    echo "! Reinstalling should error out."
    echo "$OUT"
    exit 1
fi

echo "- Uninstaller 1"
"$GH_TEST_BIN/githooks-cli" uninstaller || {
    echo "! Uninstall should have worked"
    exit 1
}

if command -v git-lfs &>/dev/null &&
    ! [ -f .git/hooks/pre-push ]; then
    echo "! Git LFS was not reinstalled."
    ls -al .git/hooks
    exit 1
fi

check_no_local_install .

echo "- Installer 2"
"$GH_TEST_BIN/githooks-cli" \
    installer \
    --prefix ~ \
    --centralized
check_centralized_install

echo "- Install local again (fail)"
OUT=$("$GH_INSTALL_BIN_DIR/githooks-cli" install --maintained-hooks "!all, pre-commit" 2>&1)
EXIT_CODE="$?"
# shellcheck disable=SC2181
if [ "$EXIT_CODE" -eq "0" ] ||
    ! echo "$OUT" | grep -qiE "Githooks is installed in 'centralized' mode and" ||
    ! echo "$OUT" | grep -qiE "installing into the current.* has no effect"; then
    echo "! Partial install with run-wrappers and global core.hooksPath should error out."
    echo "$OUT"
    exit 1
fi

echo "- Uninstaller 2"
"$GH_TEST_BIN/githooks-cli" uninstaller || {
    echo "! Uninstall should have worked"
    exit 1
}

echo "- Installer 3"
"$GH_TEST_BIN/githooks-cli" installer "${EXTRA_INSTALL_ARGS[@]}" --prefix ~

check_normal_install
check_no_local_install .

echo "- Install local partially"
"$GH_INSTALL_BIN_DIR/githooks-cli" install --maintained-hooks "!all, pre-commit" || exit 1
if command -v git-lfs &>/dev/null; then
    check_install_hooks_local . 5 "pre-commit"
else
    check_install_hooks_local . 1 "pre-commit"
fi

echo "- Check that pre-commit runs."
rm -rf "$GH_TEST_TMP/test-009.out" || true
git commit --allow-empty -m 'Test' &>/dev/null || exit 1
if ! grep 'In-repo' "$GH_TEST_TMP/test-009.out"; then
    echo '! Hooks should have been run.'
    exit 1
fi
