#!/bin/sh
# Test:
#   Cli tool: list hooks for all types of hook sources

TEST_DIR=$(cd "$(dirname "$0")" && pwd)
# shellcheck disable=SC1090
. "$TEST_DIR/general.sh"

acceptAllTrustPrompts || exit 1

"$GH_TEST_BIN/cli" installer || exit 1

url1="ssh://git@github.com/test/repo1.git"
location1=$("$GH_INSTALL_BIN_DIR/cli" shared root-from-url "$url1") || exit 1

url2="https://github.com/test/repo2.git"
location2=$("$GH_INSTALL_BIN_DIR/cli" shared root-from-url "$url2") || exit 1

url3="ftp://github.com/test/repo3.git"
location3=$("$GH_INSTALL_BIN_DIR/cli" shared root-from-url "$url3") || exit 1

# Shared with hooks in root directory.
mkdir -p "$location1"/pre-commit &&
    cd "$location1" &&
    git init &&
    git remote add origin "$url1" &&
    echo "repo1" >.namespace &&
    echo 'echo "Hello"' >pre-commit/shared-pre1 &&
    echo 'echo "Hello"' >commit-msg &&
    mkdir -p "$location1"/pre-commit/step-1 &&
    echo 'echo "Hello"' >pre-commit/step-1/step1.1 &&
    echo 'echo "Hello"' >pre-commit/step-1/step1.2 ||
    exit 1

# Shared with hooks in 'githooks' directory.
mkdir -p "$location2"/githooks/pre-push &&
    cd "$location2" &&
    git init &&
    git remote add origin "$url2" &&
    echo "repo2" >githooks/.namespace &&
    echo 'echo "Hello"' >githooks/post-commit &&
    echo 'echo "Hello"' >githooks/pre-push/shared-pre2 &&
    mkdir -p githooks/pre-push/step-2 &&
    echo 'echo "Hello"' >githooks/pre-push/step-2/step2.1 &&
    echo 'echo "Hello"' >githooks/pre-push/step-2/step2.2 ||
    exit 1

# Shared with hooks in '.githooks' directory.
mkdir -p "$location3"/.githooks/post-update &&
    cd "$location3" &&
    git init &&
    git remote add origin "$url3" &&
    echo "repo3" >.githooks/.namespace &&
    echo 'echo "Hello"' >.githooks/post-rewrite &&
    echo 'echo "Hello"' >.githooks/post-update/shared-pre3 &&
    mkdir -p .githooks/post-update/step-3 &&
    echo 'echo "Hello"' >.githooks/post-update/step-3/step3.1 &&
    echo 'echo "Hello"' >.githooks/post-update/step-3/step3.2 ||
    exit 1

mkdir -p "$GH_TEST_TMP/test055" &&
    cd "$GH_TEST_TMP/test055" &&
    mkdir -p ".githooks/pre-commit" &&
    mkdir -p ".githooks/post-commit" &&
    echo 'echo "Hello"' >.githooks/pre-commit/local-pre &&
    echo 'echo "Hello"' >.githooks/post-commit/local-post &&
    echo 'echo "Hello"' >.githooks/post-merge &&
    echo "urls: - $url2" >.githooks/.shared.yaml &&
    mkdir -p .githooks/post-commit/step-4 &&
    echo 'echo "Hello"' >.githooks/post-commit/step-4/step4.1 &&
    echo 'echo "Hello"' >.githooks/post-commit/step-4/step4.2 ||
    exit 1

cd "$GH_TEST_TMP/test055" &&
    git init &&
    mkdir -p .git/hooks &&
    echo 'echo "Hello"' >.git/hooks/pre-commit.replaced.githook &&
    chmod +x .git/hooks/pre-commit.replaced.githook &&
    git config --local githooks.shared "$url3" ||
    exit 1

git config --global githooks.shared "$url1" || exit 1

if ! "$GH_INSTALL_BIN_DIR/cli" list pre-commit | grep -q "'replaced'"; then
    echo "! Unexpected cli list output (1)"
    exit 1
fi

if ! "$GH_INSTALL_BIN_DIR/cli" list pre-commit | grep "shared-pre1" | grep -q "'shared:global'"; then
    echo "! Unexpected cli list output (2)"
    exit 1
fi

if ! "$GH_INSTALL_BIN_DIR/cli" list pre-commit | grep "local-pre" | grep "'repo'"; then
    echo "! Unexpected cli list output (3)"
    exit 1
fi

if ! "$GH_INSTALL_BIN_DIR/cli" list commit-msg | grep "'shared:global'" | grep -q "commit-msg"; then
    echo "! Unexpected cli list output (4)"
    exit 1
fi

if ! "$GH_INSTALL_BIN_DIR/cli" list post-commit | grep "local-post" | grep -q "'repo'"; then
    echo "! Unexpected cli list output (6)"
    exit 1
fi

if ! "$GH_INSTALL_BIN_DIR/cli" list post-commit | grep "'shared:repo'" | grep -q "post-commit"; then
    echo "! Unexpected cli list output (5)"
    exit 1
fi

if ! "$GH_INSTALL_BIN_DIR/cli" list post-merge | grep "'repo'" | grep -q "post-merge"; then
    echo "! Unexpected cli list output (7)"
    exit 1
fi

if ! "$GH_INSTALL_BIN_DIR/cli" list pre-push | grep "shared-pre2" | grep -q "'shared:repo'"; then
    echo "! Unexpected cli list output (8)"
    exit 1
fi

if ! "$GH_INSTALL_BIN_DIR/cli" list post-update | grep "shared-pre3" | grep -q "'shared:local'"; then
    echo "! Unexpected cli list output (9)"
    exit 1
fi

if ! "$GH_INSTALL_BIN_DIR/cli" list post-rewrite | grep "'shared:local'" | grep -q "'post-rewrite'"; then
    echo "! Unexpected cli list output (10)"
    exit 1
fi

if ! "$GH_INSTALL_BIN_DIR/cli" list | grep -q "Total.*hooks: '18'"; then
    echo "! Unexpected cli list output (12)"
    exit 1
fi

# Check all parallel batch names
OUT=$("$GH_INSTALL_BIN_DIR/cli" list --batch-name) || exit 11
if ! echo "$OUT" | grep "step1.1" | grep -q -E "ns-path: +'ns:repo1/pre-commit/step-1/step1.1'" ||
    ! echo "$OUT" | grep "step1.1" | grep -q -E "batch: +'step-1'" ||
    ! echo "$OUT" | grep "step1.2" | grep -q -E "ns-path: +'ns:repo1/pre-commit/step-1/step1.2'" ||
    ! echo "$OUT" | grep "step1.2" | grep -q -E "batch: +'step-1'"; then
    echo "! Unexpected cli list output (11):"
    echo "$OUT"
    exit 1
fi

if ! echo "$OUT" | grep "step2.1" | grep -q -E "ns-path: +'ns:repo2/pre-push/step-2/step2.1'" ||
    ! echo "$OUT" | grep "step2.1" | grep -q -E "batch: +'step-2'" ||
    ! echo "$OUT" | grep "step2.2" | grep -q -E "ns-path: +'ns:repo2/pre-push/step-2/step2.2'" ||
    ! echo "$OUT" | grep "step2.2" | grep -q -E "batch: +'step-2'"; then
    echo "! Unexpected cli list output (12):"
    echo "$OUT"
    exit 1
fi

if ! echo "$OUT" | grep "step3.1" | grep -q -E "ns-path: +'ns:repo3/post-update/step-3/step3.1'" ||
    ! echo "$OUT" | grep "step3.1" | grep -q -E "batch: +'step-3'" ||
    ! echo "$OUT" | grep "step3.2" | grep -q -E "ns-path: +'ns:repo3/post-update/step-3/step3.2'" ||
    ! echo "$OUT" | grep "step3.2" | grep -q -E "batch: +'step-3'"; then
    echo "! Unexpected cli list output (13):"
    echo "$OUT"
    exit 1
fi

if ! echo "$OUT" | grep "step4.1" | grep -q -E "ns-path: +'ns:gh-self/post-commit/step-4/step4.1'" ||
    ! echo "$OUT" | grep "step4.1" | grep -q -E "batch: +'step-4'" ||
    ! echo "$OUT" | grep "step4.2" | grep -q -E "ns-path: +'ns:gh-self/post-commit/step-4/step4.2'" ||
    ! echo "$OUT" | grep "step4.2" | grep -q -E "batch: +'step-4'"; then
    echo "! Unexpected cli list output (14):"
    echo "$OUT"
    exit 1
fi

# Check if we can get the location
root1=$(git hooks shared root "ns:repo1")
if [ "$root1" != "$location1" ]; then
    echo "! Unexpected cli shared root output (15):"
    echo "'$root1' != '$location1'"
    exit 1
fi

root2=$(git hooks shared root "ns:repo2")
if [ "$root2" != "$location2" ]; then
    echo "! Unexpected cli shared root output (16):"
    echo "'$root2' != '$location2'"

    exit 1
fi

root3=$(git hooks shared root "ns:repo3")
if [ "$root3" != "$location3" ]; then
    echo "! Unexpected cli shared root output (17):"
    echo "'$root3' != '$location3'"

    exit 1
fi
