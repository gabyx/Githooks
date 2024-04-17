#!/usr/bin/env bash
# Test:
#   Run an centralized install and check for wrong install usage

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

if ! echo "${EXTRA_INSTALL_ARGS:-}" | grep -q "centralized"; then
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

check_centralized_install

echo "- Check that pre-commit runs."
git commit -m 'Test' &>/dev/null || exit 1
if ! grep 'In-repo' "$GH_TEST_TMP/test-009.out"; then
    echo '! Hooks should have been run.'
    exit 1
fi

# Install into current repo should fail.
OUT=$("$GH_INSTALL_BIN_DIR/githooks-cli" install 2>&1)
# shellcheck disable=SC2181
if [ "$?" -eq "0" ] ||
    ! echo "$OUT" | grep -iqE "installing into the current repository has no effect"; then
    echo "! Installing into repo with centralized install must error out."
    echo "$OUT"
    exit 1
fi

echo "- Installer 1"
# Install again but with different install
# Wrap cmd on new line to prohibit regex-replacement for add arguments.
OUT=$("$GH_TEST_BIN/githooks-cli" \
    installer 2>&1)
# shellcheck disable=SC2181
if [ "$?" -eq "0" ] || ! echo "$OUT" | grep -qiE "You seem to have already installed Githooks in mode"; then
    echo "! Reinstalling should error out."
    echo "$OUT"
    exit 1
fi

# Uninstall
echo "- Uninstaller 1"
"$GH_TEST_BIN/githooks-cli" uninstaller || {
    echo "! Uninstall should have worked"
    exit 1
}
if ! [ -f .git/hooks/pre-push ]; then
    echo "! Git LFS was not reinstalled."
    ls -al .git/hooks
    exit 1
fi
check_no_local_install .

# Reinstall
echo "- Installer 2"
"$GH_TEST_BIN/githooks-cli" \
    installer \
    --prefix ~
check_normal_install

# Install into current
echo "- Install local"
"$GH_INSTALL_BIN_DIR/githooks-cli" install
check_local_install .
check_local_install_no_run_wrappers .

# Install some manual run-wrappers.
# Check for failure.
echo "- Install local partially (reject)"
git config --global core.hooksPath "/this-is-a-test"
OUT=$("$GH_INSTALL_BIN_DIR/githooks-cli" install --maintained-hooks "pre-commit" 2>&1)
EXIT_CODE="$?"
# shellcheck disable=SC2181
if [ "$EXIT_CODE" -eq "0" ] ||
    ! echo "$OUT" | grep -qiE "Global Git config 'core\.hooksPath.* is set" ||
    ! echo "$OUT" | grep -qiE "which circumvents Githooks run-wrappers"; then
    echo "! Partial install with run-wrappers and global core.hooksPath should error out."
    echo "$OUT"
    exit 1
fi
echo "- Install local partially (success)"
git config --global --unset core.hooksPath
"$GH_INSTALL_BIN_DIR/githooks-cli" install --maintained-hooks "!all, pre-commit" || exit 1
check_local_install_run_wrappers .
check_install_hooks_local . 5 "pre-commit"

# Uninstall
echo "- Uninstaller 2"
"$GH_TEST_BIN/githooks-cli" uninstaller || {
    echo "! Uninstall should have worked"
    exit 1
}
check_no_local_install .

echo "- Install 3"
# Install again which should have installed run-wrappers into registered
# repo.
"$GH_TEST_BIN/githooks-cli" installer "${EXTRA_INSTALL_ARGS[@]}" \
    --prefix ~ \
    --centralized

check_no_local_install .
