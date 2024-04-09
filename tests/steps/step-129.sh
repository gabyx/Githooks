#!/usr/bin/env bash
# Test:
#   Test maintainable hooks at install.

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

if ! command -v git-lfs; then
    echo "git-lfs is not available"
    exit 249
fi

if echo "${EXTRA_INSTALL_ARGS:-}" | grep -q "centralized"; then
    echo "Using centralized install"
    exit 249
fi

function check_lfs_hook() {
    local repo="$1"
    shift 1
    local hooks=("$@")
    for hook in "${hooks[@]}"; do
        grep -qE "^git lfs" "$repo/.git/hooks/$hook" ||
            {
                echo "Hook '$hook' should be a Git LFS hook."
                cat "$repo/.git/hooks/$hook"
                ls -al "$repo/.git/hooks"
                exit 1
            }
    done

    # All other hooks need to be Githooks
    local list
    list=$(printf '%s\n' "${hooks[@]}")
    for path in "$repo/.git/hooks"/*; do
        basename=$(basename "$path")
        if ! echo "$list" | grep -Fxq "$basename"; then

            # Skip custom hooks.
            if grep -q "custom-to-survive" "$path"; then
                continue
            fi

            # Skip disabled hooks.
            if echo "$basename" | grep -q ".disabled."; then
                continue
            fi

            grep -qE "gabyx/githooks" "$path" ||
                {
                    echo "Hook '$path' should be a Githooks run-wrapper."
                    cat "$path"
                    ls -al "$repo/.git/hooks"
                    exit 1
                }
        fi
    done
}

function check_hooks() {
    local repo="$1"
    shift 1
    local hooks=("$@")
    local count="${#hooks[@]}"

    for hook in "${hooks[@]}"; do
        [ -f "$repo/.git/hooks/$hook" ] ||
            {
                echo "Hook '$hook' should be existing in repo '$repo'"
                ls -al "$repo/.git/hooks"
                exit 1
            }
    done

    [ -n "$ADD_COUNT" ] && count=$((count + ADD_COUNT))

    countCurrent=$(find "$repo/.git/hooks" -type f -not -name "*.replaced.*" -mindepth 1 -maxdepth 1 2>/dev/null | wc -l)

    [ "$countCurrent" = "$count" ] ||
        {
            echo "Repo '$repo' should contain '$count' hooks (current: '$countCurrent')."
            ls -al "$repo/.git/hooks"
            exit 1
        }

}

allLFSHooks=(
    "post-checkout"
    "post-commit"
    "post-merge"
    "pre-push")

echo "Install maintained hooks."
maintainedHooks1="!all,  commit-msg, post-applypatch,  post-checkout"
maintainedHooksRef1=(
    "commit-msg"
    "post-applypatch"
    "post-checkout"
    "post-commit"
    "post-merge"
    "pre-push")
lfsHooks1=(
    "post-commit"
    "post-merge"
    "pre-push")

mkdir -p ~/.githooks/templates
"$GH_TEST_BIN/githooks-cli" installer --maintained-hooks "$maintainedHooks1" --template-dir ~/.githooks/templates || exit 1
accept_all_trust_prompts || exit 1

[ -n "$(git config --global githooks.maintainedHooks)" ] || {
    echo "Global maintained hooks should be set, but it is not."
    exit 1
}

# Check init works
mkdir -p "$GH_TEST_TMP/test129" &&
    cd "$GH_TEST_TMP/test129" &&
    git init &&
    cd "$GH_TEST_TMP/test129" || exit 1

if [ -n "$(git config githooks.registered)" ]; then
    echo "Repository should not be registered."
    exit 1
fi

check_hooks "." "${maintainedHooksRef1[@]}"
check_lfs_hook "." "${lfsHooks1[@]}"

echo " Change maintainable hooks, locally."
maintainedHooks2="!all,  commit-msg"
maintainedHooksRef2=(
    "commit-msg"
    "post-checkout"
    "post-commit"
    "post-merge"
    "pre-push")

"$GH_TEST_BIN/githooks-cli" install --maintained-hooks "$maintainedHooks2" || exit 1
if [ "$(git config githooks.maintainedHooks)" != "!all, commit-msg" ]; then
    echo "Maintained hooks is not set"
    exit 1
fi

if [ "$(git config githooks.registered)" != "true" ]; then
    echo "Repository should be registered."
    exit 1
fi

check_hooks "." "${maintainedHooksRef2[@]}"
check_lfs_hook "." "${allLFSHooks[@]}"

echo "Uninstall and place a custom hook which should survive."
"$GH_TEST_BIN/githooks-cli" uninstall || exit 1
check_hooks "." "${allLFSHooks[@]}"
check_lfs_hook "." "${allLFSHooks[@]}"
echo "echo 'custom-to-survive'" >.git/hooks/commit-msg

echo "Change maintainable hooks again."
maintainedHooks3="!all, post-checkout,  post-commit,  post-merge,pre-push"
maintainedHooksRef3=(
    "post-checkout"
    "post-commit"
    "post-merge"
    "pre-push")
lfsHooks3=()

"$GH_TEST_BIN/githooks-cli" install --maintained-hooks "$maintainedHooks3" || exit 1
echo "Delete disabled LFS hooks."
find .git/hooks -type f -name "*.disabled.*" -exec rm -f {} \; || exit 1
grep -q "custom-to-survive" .git/hooks/commit-msg || {
    echo "Replaced hook should still exist."
    ls -al .git/hooks
    cat .git/hooks/commit-msg
    exit 1
}

export ADD_COUNT=1 # commit-msg
check_hooks "." "${maintainedHooksRef3[@]}"
check_lfs_hook "." "${lfsHooks3[@]}"

echo "Change maintainable hooks, locally. again"
maintainedHooks4="!all, server"
maintainedHooksRef4=(
    "pre-push"
    "pre-receive"
    "update"
    "post-receive"
    "post-update"
    "reference-transaction"
    "push-to-checkout"
    "pre-auto-gc"
    #LFS Hooks
    "post-checkout"
    "post-commit"
    "post-merge")
lfsHooks4=(
    "post-checkout"
    "post-commit"
    "post-merge")

git config githooks.maintainedHooks "$maintainedHooks4"
"$GH_TEST_BIN/githooks-cli" install || exit 1

check_hooks "." "${maintainedHooksRef4[@]}"
check_lfs_hook "." "${lfsHooks4[@]}"

echo "Pollute an LFS hook and reinstall again."
echo "echo 'overwritten LFS hooks'" >.git/hooks/post-checkout
"$GH_TEST_BIN/githooks-cli" install && {
    echo "'git hooks install' should have failed, because cannot overwrite existing LFS hook."
    exit 1
}
grep -q "overwritten LFS hooks" .git/hooks/post-checkout || {
    echo "Pollute LFS hooks should have stayed the same but did not."
    cat .git/hooks/post-checkout
    exit 1
}

echo "Delete the polluted LFS hook an run again"
rm .git/hooks/post-checkout || exit 1
"$GH_TEST_BIN/githooks-cli" install || exit 1
check_hooks "." "${maintainedHooksRef4[@]}"
check_lfs_hook "." "${lfsHooks4[@]}"

echo "Uninstall all hooks, check that all LFS hooks are installed."
"$GH_TEST_BIN/githooks-cli" uninstall || exit 1
check_hooks "." "${allLFSHooks[@]}"
check_lfs_hook "." "${allLFSHooks[@]}"

echo "Unset git config githooks.maintainedHooks and check that original setup is maintained."
git config --unset githooks.maintainedHooks
"$GH_TEST_BIN/githooks-cli" install || exit 1
check_hooks "." "${maintainedHooksRef1[@]}"
check_lfs_hook "." "${lfsHooks1[@]}"
grep -q "custom-to-survive" .git/hooks/commit-msg.replaced.githook || {
    echo "Replaced hook should still exist."
    ls -al .git/hooks/
    cat .git/hooks/commit-msg.replaced.githook
    exit 1
}
