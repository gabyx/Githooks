#!/usr/bin/env bash
# Test:
#   Set up bare repos, run the install and verify the hooks get installed/uninstalled

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

if echo "${EXTRA_INSTALL_ARGS:-}" | grep -q "centralized"; then
    echo "Using centralized install"
    exit 249
fi

accept_all_trust_prompts || exit 1

mkdir -p "$GH_TEST_TMP/test109/p001" &&
    mkdir -p "$GH_TEST_TMP/test109/p002" &&
    mkdir -p "$GH_TEST_TMP/test109/p003" || exit 1

cd "$GH_TEST_TMP/test109/p001" && git init --bare || exit 1
cd "$GH_TEST_TMP/test109/p002" && git init --bare || exit 1

check_no_local_install "$GH_TEST_TMP/test109/p001"
check_no_local_install "$GH_TEST_TMP/test109/p002"

mkdir -p "$GH_TEST_TMP/.githooks/templates/hooks"
git config --global init.templateDir "$GH_TEST_TMP/.githooks/templates"

# run the install, and select installing hooks into existing repos
echo "n
y
$GH_TEST_TMP/test109
" | "$GH_TEST_BIN/githooks-cli" installer "${EXTRA_INSTALL_ARGS[@]}" --hooks-dir-use-template-dir \
    --maintained-hooks server --stdin || exit 1

check_local_install "$GH_TEST_TMP/test109/p001"
check_local_install "$GH_TEST_TMP/test109/p002"

# check if only server hooks are installed.
# + 3 missing LFS hooks (are always installed due to safety, also not needed)
if command -v git-lfs; then
    check_install_hooks \
        11 \
        pre-push pre-receive update post-receive post-update push-to-checkout pre-auto-gc
else
    check_install_hooks \
        8 \
        pre-push pre-receive update post-receive post-update push-to-checkout pre-auto-gc
fi

cd "$GH_TEST_TMP/test109/p003" &&
    git init --bare || exit 1

# we should have run-wrappers installed due to template dir above.
check_local_install_run_wrappers "$GH_TEST_TMP/test109/p003"
"$GH_INSTALL_BIN_DIR/githooks-cli" install
# Install should not have changed (still runwrappers)
check_local_install_run_wrappers "$GH_TEST_TMP/test109/p003"

echo "y
$GH_TEST_TMP/test109
" | "$GH_TEST_BIN/githooks-cli" uninstaller --stdin || exit 1

check_no_local_install "$GH_TEST_TMP/test109/p001"
check_no_local_install "$GH_TEST_TMP/test109/p002"
check_no_local_install "$GH_TEST_TMP/test109/p003"
