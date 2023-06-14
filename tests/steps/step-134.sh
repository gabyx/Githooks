#!/usr/bin/env bash
# Test:
#   Update shared hooks with images.yaml
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
    echo "sharedhooks" >"githooks/.namespace" &&
    git add . &&
    git commit -m 'Initial commit' ||
    exit 1

# Setup local repository
mkdir -p "$GH_TEST_TMP/test134" &&
    cd "$GH_TEST_TMP/test134" &&
    mkdir -p .githooks &&
    echo -e "urls:\n  - file://$GH_TEST_TMP/shared/hooks-134-a.git" >.githooks/.shared.yaml &&
    git init &&
    git add . &&
    GITHOOKS_DISABLE=1 git commit -m 'Initial commit' ||
    exit 1

"$GH_TEST_BIN/cli" shared update
# "$GH_TEST_BIN/cli" images update

touch "file.txt" &&
    git add "file.txt"

# Creating volumes for the mounting, because
# `docker in docker` uses directories on host volume.
sharedRoot=$("$GH_TEST_BIN/cli" shared root ns:sharedhooks)
storeIntoContainerVolumes "." "$sharedRoot"
OUT=$(git commit -m "fix: Add file to format")
restoreFromContainerVolumeWorkspace "." "file.txt"

echo "$OUT"
if ! echo "$OUT" | grep -iq "formatting file 'file.txt'"; then
    echo -e "! Expected file to have formatted"
    exit 1
fi

if ! grep -qi "formatted by containerized hook" "file.txt"; then
    echo -e "! Expected file should have been changed"
    exit 1
fi

deleteAllTestImages
