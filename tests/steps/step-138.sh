#!/usr/bin/env bash
# Test:
#   Run CLI exec.
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

mkdir -p "$GH_TEST_TMP/shared/hooks-138-a.git" &&
    cd "$GH_TEST_TMP/shared/hooks-138-a.git" &&
    git init &&
    mkdir githooks &&
    cp -rf "$TEST_DIR/steps/images/image-1/.images.yaml" ./githooks/.images.yaml &&
    cp -rf "$TEST_DIR/steps/images/image-1/docker" ./docker &&
    cp -rf "$TEST_DIR/steps/images/image-1/githooks/scripts" githooks/scripts &&
    echo "sharedhooks" >"githooks/.namespace" &&
    git add . &&
    git commit -m 'Initial commit' ||
    exit 1

# Setup local repository
mkdir -p "$GH_TEST_TMP/test138" &&
    cd "$GH_TEST_TMP/test138" &&
    git init &&
    mkdir -p .githooks &&
    echo -e "urls:\n  - file://$GH_TEST_TMP/shared/hooks-138-a.git" >.githooks/.shared.yaml &&
    GITHOOKS_DISABLE=1 git add . &&
    GITHOOKS_DISABLE=1 git commit -m 'Initial commit' ||
    exit 1

# Enable containerized hooks.
export GITHOOKS_CONTAINERIZED_HOOKS_ENABLED=true

"$GH_TEST_BIN/cli" shared update

# Creating volumes for the mounting, because
# `docker in docker` uses directories on host volume,
# which we dont have.
storeIntoContainerVolumes "." "$HOME/.githooks/shared"
showAllContainerVolumes 3

OUT=$(setGithooksContainerVolumeEnvs &&
    git hooks exec ns:sharedhooks/scripts/test-success.sh "arg1" "arg2" 2>&1) ||
    {
        echo "Execution failed."
        echo "$OUT"
        exit 1
    }

if ! echo "$OUT" | grep -iq "executing test script 'arg1' 'arg2'"; then
    echo "! Expected output not found."
    exit 1
fi

OUT=$(setGithooksContainerVolumeEnvs &&
    git hooks exec ns:sharedhooks/scripts/test-fail.sh 2>&1)

if [ "$?" != "123" ]; then
    echo "! Test script should have reported 123."
    exit 1
fi

if ! echo "$OUT" | grep -iq "executing test script"; then
    echo "! Expected output not found."
    exit 1
fi

deleteContainerVolumes
deleteAllTestImages
