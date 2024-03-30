#!/usr/bin/env bash
# Test:
#   Git LFS integration

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

accept_all_trust_prompts || exit 1

if ! command -v git-lfs; then
    echo "git-lfs is not available"
    exit 249
fi

mkdir -p "$GH_TEST_TMP/test105" &&
    cd "$GH_TEST_TMP/test105" &&
    git init &&
    git lfs install ||
    exit 1

IFS="
"

LFS_UNMANAGED=""

# shellcheck disable=SC2013
for LFS_HOOK_PATH in $(grep -l git-lfs .git/hooks/*); do
    LFS_HOOK=$(basename "$LFS_HOOK_PATH")

    if ! sed -n -E '/LFSHookNames\s*=.*\{/,/\}/p;' "$GH_TEST_REPO/githooks/hooks/githooks.go" | grep -q "$LFS_HOOK"; then
        echo "! LFS hook appears unmanaged: $LFS_HOOK"
        LFS_UNMANAGED=Y
    fi
done

unset IFS

[ -z "$LFS_UNMANAGED" ] || exit 2
