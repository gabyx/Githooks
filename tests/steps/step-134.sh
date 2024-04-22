#!/usr/bin/env bash
# Test:
#   Run shared hooks with images.yaml
set -e
set -u

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

init_step

# Test can be run with staged files exported as file too.
exportStagedFilesAsFile="false"
if [ "${1:-}" = "--export-staged-files-as-file" ]; then
    exportStagedFilesAsFile="true"
fi

if ! is_docker_available; then
    echo "docker is not available"
    exit 249
fi

"$GH_TEST_BIN/githooks-cli" installer "${EXTRA_INSTALL_ARGS[@]}" || exit 1

accept_all_trust_prompts || exit 1
assert_no_test_images

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
    git commit -m 'Initial commit'

# Setup local repository
mkdir -p "$GH_TEST_TMP/test134" &&
    cd "$GH_TEST_TMP/test134" &&
    git init &&
    install_hooks_if_not_centralized &&
    mkdir -p .githooks &&
    echo -e "envs:\n  sharedhooks:\n    - MONKEY=gaga" >.githooks/.envs.yaml &&
    echo -e "urls:\n  - file://$GH_TEST_TMP/shared/hooks-134-a.git" >.githooks/.shared.yaml

# Maybe run with exported staged files file.
if [ "$exportStagedFilesAsFile" = "true" ]; then
    git config --global githooks.exportStagedFilesAsFile true
    touch .githooks-test-export-staged-files
fi

# Commit all files.
GITHOOKS_DISABLE=1 git add . &&
    GITHOOKS_DISABLE=1 git commit -m 'Initial commit'

# Enable containerized hooks.
export GITHOOKS_CONTAINERIZED_HOOKS_ENABLED=true

"$GH_TEST_BIN/githooks-cli" shared update
# The above already does the below
# "$GH_TEST_BIN/githooks-cli" images update

# Make changes to be formatted.
touch "file.txt" &&
    touch "file-2.txt" &&
    GITHOOKS_DISABLE=1 git add .

# Creating volumes for the mounting, because
# `docker in docker` uses directories on host volume,
# which we dont have.
store_into_container_volumes "$HOME/.githooks/shared"
show_all_container_volumes 3
set_githooks_container_volume_envs "."

OUT=$(git commit -m "fix: Add file to format" 2>&1) ||
    {
        echo "! Commit failed"
        echo "$OUT"
        exit 1
    }

echo "$OUT"

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

if [ "$(find .githooks -name ".githooks-staged.*" | wc -l)" != "0" ]; then
    echo "Staged file still exists in .githooks."
    exit 1
fi

delete_container_volumes
delete_all_test_images
