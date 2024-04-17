#!/usr/bin/env bash
# Test:
#   Fail on not available shared hooks.

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

init_step

accept_all_trust_prompts || exit 1

git config --global githooks.testingTreatFileProtocolAsRemote "true"

if ! "$GH_TEST_BIN/githooks-cli" installer "${EXTRA_INSTALL_ARGS[@]}"; then
    echo "! Failed to execute the install script"
    exit 1
fi

mkdir -p "$GH_TEST_TMP/shared/hooks-103.git/pre-commit" &&
    echo 'exit 0' >"$GH_TEST_TMP/shared/hooks-103.git/pre-commit/succeed" &&
    cd "$GH_TEST_TMP/shared/hooks-103.git" &&
    git init &&
    git add . &&
    git commit -m 'Initial commit' ||
    exit 1

# Install shared hook url into a repo.
mkdir -p "$GH_TEST_TMP/test103" &&
    cd "$GH_TEST_TMP/test103" &&
    git init &&
    install_hooks_if_not_centralized || exit 1

mkdir -p .githooks && echo "urls: - file://$GH_TEST_TMP/shared/hooks-103.git" >.githooks/.shared.yaml || exit 1
git add .githooks/.shared.yaml
"$GH_INSTALL_BIN_DIR/githooks-cli" shared update

# shellcheck disable=SC2012
RESULT=$(find ~/.githooks/shared/ -type f 2>/dev/null | wc -l)
if [ "$RESULT" = "0" ]; then
    echo "! Expected shared hooks to be installed."
    exit 1
fi

git commit -m "Test" || exit 1

# Remove all shared hooks and make it fail
"$GH_INSTALL_BIN_DIR/githooks-cli" shared purge || exit 1

if [ -d ~/.githooks/shared ]; then
    echo "! Expected shared hooks to be purged."
    exit 1
fi

# Test some random nonsense.
! "$GH_INSTALL_BIN_DIR/githooks-cli" config skip-non-existing-shared-hooks --enable --disable || exit 1
! "$GH_INSTALL_BIN_DIR/githooks-cli" config skip-non-existing-shared-hooks --enable --print || exit 1
! "$GH_INSTALL_BIN_DIR/githooks-cli" config skip-non-existing-shared-hooks --disable --print || exit 1
! "$GH_INSTALL_BIN_DIR/githooks-cli" config skip-non-existing-shared-hooks --local --global --enable || exit 1

# Skip on not existing hooks
# Local off/ global on
if ! "$GH_INSTALL_BIN_DIR/githooks-cli" config skip-non-existing-shared-hooks --disable; then
    echo "! Disabling skip-non-existing-shared-hooks failed"
    exit 1
fi

if ! "$GH_INSTALL_BIN_DIR/githooks-cli" config skip-non-existing-shared-hooks --local --print | grep -q "disabled"; then
    echo "! Expected skip-non-existing-shared-hooks to be disabled locally"
    exit 1
fi

if ! "$GH_INSTALL_BIN_DIR/githooks-cli" config skip-non-existing-shared-hooks --global --print | grep -q "disabled"; then
    echo "! Expected skip-non-existing-shared-hooks to be disabled globally"
    exit 1
fi

if [ ! "$(git config --local --get githooks.skipNonExistingSharedHooks)" = "false" ]; then
    echo "! Expected githooks.skipNonExistingSharedHooks to be disabled locally"
    exit 1
fi

if git config --global --get githooks.skipNonExistingSharedHooks; then
    echo "! Expected githooks.skipNonExistingSharedHooks to be unset globally"
    exit 1
fi

# Local off / global off
if ! "$GH_INSTALL_BIN_DIR/githooks-cli" config skip-non-existing-shared-hooks --global --disable; then
    echo "! Disabling skip-non-existing-shared-hooks globally failed"
    exit 1
fi

if ! "$GH_INSTALL_BIN_DIR/githooks-cli" config skip-non-existing-shared-hooks --global --print | grep -q "disabled"; then
    echo "! Expected skip-non-existing-shared-hooks to be disabled globally"
    exit 1
fi

if [ ! "$(git config --local --get githooks.skipNonExistingSharedHooks)" = "false" ]; then
    echo "! Expected githooks.skipNonExistingSharedHooks to be still disabled locally"
    exit 1
fi

if [ ! "$(git config --global --get githooks.skipNonExistingSharedHooks)" = "false" ]; then
    echo "! Expected githooks.skipNonExistingSharedHooks to be set globally"
    exit 1
fi

# Clone a new one, set template dir
echo "Cloning"
cd "$GH_TEST_TMP" || exit 1

if ! is_centralized_tests; then
    # Set templateDir such that shared hooks update gets triggered on clone.
    git -c "init.templateDir=$(git config githooks.pathForUseCoreHooksPath)/.." \
        clone "$GH_TEST_TMP/test103" test103-clone &&
        cd test103-clone || exit 1
else
    git clone "$GH_TEST_TMP/test103" test103-clone &&
        cd test103-clone || exit 1
fi

# shellcheck disable=SC2012
RESULT=$(find ~/.githooks/shared/ -type f 2>/dev/null | wc -l)
if [ "$RESULT" = "0" ]; then
    echo "! Expected shared hooks to be installed."
    exit 1
fi

# Remove all shared hooks
"$GH_INSTALL_BIN_DIR/githooks-cli" shared purge || exit 1

echo "Committing"
# Make a commit
echo A >A || exit 1
git add A || exit 1
OUTPUT=$(git commit -a -m "Test" 2>&1)

# shellcheck disable=SC2181
if [ $? -eq 0 ] || ! echo "$OUTPUT" | grep -q "needs shared hooks in:"; then
    echo "! Expected to fail on not available shared hooks. output:"
    echo "$OUTPUT"
    exit 1
fi

# Try to skip them.
OUTPUT=$(GITHOOKS_SKIP_NON_EXISTING_SHARED_HOOKS=true git commit -a -m "Test" 2>&1)

# shellcheck disable=SC2181
if [ $? -ne 0 ]; then
    echo "! Expected to skip on not available shared hooks. output:"
    echo "$OUTPUT"
    exit 1
fi

"$GH_INSTALL_BIN_DIR/githooks-cli" shared pull || exit 1

# Change url and try to make it fail
(cd ~/.githooks/shared/*shared-hooks-103* &&
    git remote rm origin &&
    git remote add origin /some/other/url.git) || exit 1
# Make a commit
echo A >>A || exit 1
OUTPUT=$(git commit -a -m "Test" 2>&1)

# shellcheck disable=SC2181
if [ $? -eq 0 ] || ! (echo "$OUTPUT" | grep "The remote" | grep -q "is different"); then
    echo "! Expected to fail on not matching url. output:"
    echo "$OUTPUT"
    exit 1
fi
