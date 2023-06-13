#!/usr/bin/env bash
# Test:
#   Update shared hooks with images.yaml

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

if ! isDockerAvailable; then
    echo "docker is not available"
    exit 249
fi

acceptAllTrustPrompts || exit 1
assertNoTestImages

git config --global githooks.testingTreatFileProtocolAsRemote "true"

mkdir -p "$GH_TEST_TMP/shared/hooks-132-a.git" &&
    cd "$GH_TEST_TMP/shared/hooks-132-a.git" &&
    git init &&
    mkdir githooks &&
    cp -rf "$TEST_DIR/steps/images/image-1/.images.yaml" ./githooks/.images.yaml &&
    cp -rf "$TEST_DIR/steps/images/image-1/docker" ./docker &&
    echo "sharedhooks" >"githooks/.namespace" &&
    git add . &&
    git commit -m 'Initial commit' ||
    exit 1

mkdir -p "$GH_TEST_TMP/test132" &&
    cd "$GH_TEST_TMP/test132" &&
    git init || exit 1

# Setup shared hooks
mkdir -p .githooks &&
    echo -e "urls:\n  - file://$GH_TEST_TMP/shared/hooks-132-a.git" >.githooks/.shared.yaml ||
    exit 1

# Trigger the shared hooks update
OUT=$("$GH_TEST_BIN/runner" "$(pwd)"/.git/hooks/post-merge unused 2>&1)
echo "$OUT"
if ! echo "$OUT" | grep -q "Updating shared hooks from" ||
    ! echo "$OUT" | grep -q "Build/pull images for repository"; then
    echo -e "! Expected shared hooks update"
    exit 1
fi

if ! isImageExisting "sharedhooks-test-image:1.0.0" ||
    ! isImageExisting "registry.com/sharedhooks-test-image:1.0.0" ||
    ! isImageExisting "registry.com/dir/sharedhooks-test-image-built:1.0.0"; then
    echo "Could not find all updated images."
    docker images
    exit 1
fi

deleteAllTestImages
