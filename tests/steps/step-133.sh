#!/usr/bin/env bash
# Test:
#   Update shared hooks with images.yaml

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

init_step

if ! is_docker_available; then
    echo "docker is not available"
    exit 249
fi

accept_all_trust_prompts || exit 1
assert_no_test_images

git config --global githooks.testingTreatFileProtocolAsRemote "true"

mkdir -p "$GH_TEST_TMP/shared/hooks-133-a.git" &&
    cd "$GH_TEST_TMP/shared/hooks-133-a.git" &&
    git init &&
    mkdir githooks &&
    cp -rf "$TEST_DIR/steps/images/image-1/.images.yaml" ./githooks/.images.yaml &&
    cp -rf "$TEST_DIR/steps/images/image-1/docker" ./docker &&
    echo "sharedhooks" >"githooks/.namespace" &&
    git add . &&
    git commit -m 'Initial commit' ||
    exit 1

mkdir -p "$GH_TEST_TMP/test133" &&
    cd "$GH_TEST_TMP/test133" &&
    mkdir .githooks &&
    cp -rf "$TEST_DIR/steps/images/image-1/.images.yaml" ./.githooks/.images.yaml &&
    cp -rf "$TEST_DIR/steps/images/image-1/docker" ./docker &&
    echo "localhooks" >".githooks/.namespace" &&
    echo "localhooks" >".githooks/.namespace" &&
    git init &&
    git config --local githooks.containerizedHooksEnabled true || exit 1

# Setup shared hooks
mkdir -p .githooks &&
    echo -e "urls:\n  - file://$GH_TEST_TMP/shared/hooks-133-a.git" >.githooks/.shared.yaml ||
    exit 1

"$GH_TEST_BIN/githooks-cli" shared update

if ! is_image_existing "sharedhooks-test-image:1.0.0" ||
    ! is_image_existing "registry.com/sharedhooks-test-image:1.0.0" ||
    ! is_image_existing "registry.com/dir/sharedhooks-test-image-built:1.0.0"; then
    echo "Could not find all shared images."
    docker images
    exit 1
fi

if is_image_existing "localhooks-test-image:1.0.0" ||
    is_image_existing "registry.com/localhooks-test-image:1.0.0" ||
    is_image_existing "registry.com/dir/localhooks-test-image-built:1.0.0"; then
    echo "Local images should not be build."
    docker images
    exit 1
fi

delete_all_test_images

if docker images | grep -q "test-image"; then
    echo "Could not delete all images"
    exit 1
fi

"$GH_TEST_BIN/githooks-cli" images update

if ! is_image_existing "sharedhooks-test-image:1.0.0" ||
    ! is_image_existing "registry.com/sharedhooks-test-image:1.0.0" ||
    ! is_image_existing "registry.com/dir/sharedhooks-test-image-built:1.0.0"; then
    echo "Could not find all shared images."
    docker images
    exit 1
fi

if ! is_image_existing "localhooks-test-image:1.0.0" ||
    ! is_image_existing "registry.com/localhooks-test-image:1.0.0" ||
    ! is_image_existing "registry.com/dir/localhooks-test-image-built:1.0.0"; then
    echo "Could not find all local images."
    docker images
    exit 1
fi

delete_all_test_images
