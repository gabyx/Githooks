#!/usr/bin/env bash
# Test:
#   Run an install that deletes/backups existing detected LFS hooks in existing repos

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

init_step

accept_all_trust_prompts || exit 1

if is_centralized_tests; then
    echo "Using centralized install"
    exit 249
fi

mkdir -p "$GH_TEST_TMP/test109.1/.githooks/pre-commit" &&
    cd "$GH_TEST_TMP/test109.1" &&
    echo "echo 'In-repo' >> '$GH_TEST_TMP/test-109.out'" >.githooks/pre-commit/test &&
    git init && mkdir -p .git/hooks &&
    # simulate that this repo contains wrappers, such that the install
    # installs run-wrappers below.
    touch .git/hooks/githooks-contains-run-wrappers &&
    echo "echo 'Previous1' >> '$GH_TEST_TMP/test-109.out' ; # git lfs arg1 arg2" >.git/hooks/pre-commit &&
    chmod +x .git/hooks/pre-commit ||
    exit 1

mkdir -p "$GH_TEST_TMP/test109.2/.githooks/pre-commit" &&
    cd "$GH_TEST_TMP/test109.2" && git init && mkdir -p .git/hooks &&
    # simulate that this repo contains wrappers, such that the install
    # installs run-wrappers below.
    touch .git/hooks/githooks-contains-run-wrappers &&
    echo "echo 'Previous2' >> '$GH_TEST_TMP/test-109.out' ; # git-lfs arg1 arg2" >.git/hooks/pre-commit &&
    chmod +x .git/hooks/pre-commit ||
    exit 1

mkdir -p "$GH_TEST_TMP/test109.3/.githooks/pre-commit" &&
    cd "$GH_TEST_TMP/test109.3" && git init && mkdir -p .git/hooks &&
    # simulate that this repo contains wrappers, such that the install
    # installs run-wrappers below.
    touch .git/hooks/githooks-contains-run-wrappers &&
    echo "echo 'Previous3' >> '$GH_TEST_TMP/test-109.out' ; # git  lfs arg1 arg2" >.git/hooks/pre-commit &&
    chmod +x .git/hooks/pre-commit ||
    exit 1

echo "y

n
y
$GH_TEST_TMP
y

s
" | "$GH_TEST_BIN/githooks-cli" installer "${EXTRA_INSTALL_ARGS[@]}" --stdin || exit 1

if [ -f "$GH_TEST_TMP/test109.1/.git/hooks/pre-commit.disabled.githooks" ]; then
    echo '! Expected hook to be deleted (1)'
    exit 1
fi
if [ ! -f "$GH_TEST_TMP/test109.2/.git/hooks/pre-commit.disabled.githooks" ] &&
    [ ! -f "$GH_TEST_TMP/test109.3/.git/hooks/pre-commit.disabled.githooks" ]; then
    echo '! Expected hook to be moved (2)'
    exit 1
fi

cd "$GH_TEST_TMP/test109.2" &&
    git commit --allow-empty -m 'Init' 2>/dev/null || exit 1
if grep 'Previous2' "$GH_TEST_TMP/test-109.out"; then
    echo '! Expected hook to be disabled (3)'
    exit 1
fi

cd "$GH_TEST_TMP/test109.3" &&
    git commit --allow-empty -m 'Init' 2>/dev/null || exit 1
if grep 'Previous3' "$GH_TEST_TMP/test-109.out"; then
    echo '! Expected hook to be disabled (4)'
    exit 1
fi

out=$("$GH_INSTALL_BIN_DIR/githooks-cli" config delete-detected-lfs-hooks --print)
if ! echo "$out" | grep -q "but instead backed up"; then
    echo "! Expected the correct config behavior (5)"
    echo "$out"
    exit 1
fi

# For coverage
"$GH_INSTALL_BIN_DIR/githooks-cli" config delete-detected-lfs-hooks --reset || exit 1
out=$("$GH_INSTALL_BIN_DIR/githooks-cli" config delete-detected-lfs-hooks --print)
if ! echo "$out" | grep -q "left to the user"; then
    echo "! Expected the correct config behavior (6)"
    echo "$out"
    exit 1
fi

# Reset every repo and do again
# Repo 1 no delete
# Repo 2,3 always delete
cd "$GH_TEST_TMP/test109.2/.git/hooks" && mv -f pre-commit.disabled.githooks pre-commit || exit 1
cd "$GH_TEST_TMP/test109.3/.git/hooks" && mv -f pre-commit.disabled.githooks pre-commit || exit 1
cd "$GH_TEST_TMP/test109.1" &&
    echo "echo 'Previous1' >> '$GH_TEST_TMP/test-109.out'; # git lfs arg1 arg2" >.git/hooks/pre-commit || exit 1

echo "n
y
$GH_TEST_TMP
N

a
" | "$GH_TEST_BIN/githooks-cli" installer "${EXTRA_INSTALL_ARGS[@]}" --stdin || exit 1

if [ ! -f "$GH_TEST_TMP/test109.1/.git/hooks/pre-commit.disabled.githooks" ]; then
    echo '! Expected hook to be moved (7)'
    exit 1
fi
if [ -f "$GH_TEST_TMP/test109.2/.git/hooks/pre-commit.disabled.githooks" ] &&
    [ -f "$GH_TEST_TMP/test109.3/.git/hooks/pre-commit.disabled.githooks" ]; then
    echo '! Expected hook to be deleted (8)'
    exit 1
fi
