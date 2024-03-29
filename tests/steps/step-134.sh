#!/usr/bin/env bash
# Test:
#   Run shared hooks with images.yaml
set -e
set -u

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

if ! isDockerAvailable; then
    echo "docker is not available"
    exit 249
fi

"$GH_TEST_BIN/cli" installer || exit 1

acceptAllTrustPrompts || exit 1
assertNoTestImages

git config --global githooks.testingTreatFileProtocolAsRemote "true"

mkdir -p "$GH_TEST_TMP/shared/hooks-134-a.git" &&
    cd "$GH_TEST_TMP/shared/hooks-134-a.git" &&
    git init &&
    mkdir githooks &&
    cp -rf "$TEST_DIR/steps/images/image-1/.images.yaml" ./githooks/.images.yaml &&
    cp -rf "$TEST_DIR/steps/images/image-1/docker" ./docker &&
    cp -rf "$TEST_DIR/steps/images/image-1/githooks/pre-commit" githooks/pre-commit &&
    cp -rf "$TEST_DIR/steps/images/image-1/githooks/commit-msg" githooks/commit-msg &&
    echo "sharedhooks" >"githooks/.namespace" &&
    git add . &&
    git commit -m 'Initial commit' ||
    exit 1

# Setup local repository
mkdir -p "$GH_TEST_TMP/test134" &&
    cd "$GH_TEST_TMP/test134" &&
    git init &&
    mkdir -p .githooks &&
    echo -e "urls:\n  - file://$GH_TEST_TMP/shared/hooks-134-a.git" >.githooks/.shared.yaml &&
    GITHOOKS_DISABLE=1 git add . &&
    GITHOOKS_DISABLE=1 git commit -m 'Initial commit' ||
    exit 1

# Enable containerized hooks.
export GITHOOKS_CONTAINERIZED_HOOKS_ENABLED=true

"$GH_TEST_BIN/cli" shared update
# The above already does the below
# "$GH_TEST_BIN/cli" images update

# Make changes to be formatted.
touch "file.txt" &&
    touch "file-2.txt" &&
    GITHOOKS_DISABLE=1 git add .

# Creating volumes for the mounting, because
# `docker in docker` uses directories on host volume,
# which we dont have.
storeIntoContainerVolumes "." "$HOME/.githooks/shared"
showAllContainerVolumes 3
setGithooksContainerVolumeEnvs

OUT=$(git commit -m "fix: Add file to format" 2>&1) ||
    {
        echo "! Commit failed"
        echo "$OUT"
        exit 1
    }

echo "$OUT"

restoreFromContainerVolumeWorkspace "." "file.txt" "file-2.txt" ".commit-msg-hook-run"

if [ ! -f ".commit-msg-hook-run" ]; then
    echo -e "! Expected commit-msg hook to have been run."
    exit 1
fi

if ! echo "$OUT" | grep -iq "formatting file 'file.txt'" ||
    ! echo "$OUT" | grep -iq "formatting file 'file-2.txt'"; then
    echo "! Expected file to have formatted"
    exit 1
fi

if [ "$(grep -ic "formatted by containerized hook" "file.txt")" != "1" ]; then
    echo -e "! Expected file should have been changed correctly: Content:"
    cat "file.txt"
    exit 1
fi

if [ "$(grep -ic "formatted by containerized hook" "file-2.txt")" != "1" ]; then
    echo -e "! Expected file should have been changed correctly: Content:"
    cat "file.txt"
    exit 1
fi

# Do it again, but check if staged files work too.
git config githooks.exportStagedFilesAsFile true
OUT=$(git commit -m "fix: Add file to format, with staged files file" 2>&1) ||
    {
        echo "! Commit failed"
        echo "$OUT"
        exit 1
    }

if ! echo "$OUT" | grep -iq "formatting file 'file.txt'" ||
    ! echo "$OUT" | grep -iq "formatting file 'file-2.txt'"; then
    echo "! Expected file to have formatted"
    exit 1
fi

deleteContainerVolumes
deleteAllTestImages
