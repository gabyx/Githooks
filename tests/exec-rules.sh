#!/usr/bin/env bash
# shellcheck disable=SC2015

set -e
set -u

if [ "${DOCKER_RUNNING:-}" != "true" ]; then
    echo "! This script is only meant to be run in a Docker container"
    exit 1
fi

DIR="$(cd "$(dirname "$0")" >/dev/null 2>&1 && pwd)"
REPO_DIR="$DIR/.."
REPO_DIR_ORIG="$REPO_DIR"
temp=""

# shellcheck disable=SC1091
. "$DIR/general.sh"

[ "${GH_SHOW_DIFFS:-false}" == "false" ] || echo "INFO: SHOWING DIFFS"
trap cleanUp EXIT

# shellcheck disable=SC2317
function cleanUp() {
    deleteContainerVolumes || true

    if [ -d "$temp" ]; then
        rm -rf "$temp" || true
    fi
}

function installGithooks() {
    just build &&
        "$REPO_DIR/githooks/bin/cli" installer --non-interactive --build-from-source --clone-url "file://$REPO_DIR" &&
        git clean -fX &&
        git hooks config trust-all --accept &&
        git hooks config enable-containerized-hooks --set &&
        git hooks shared update &&
        git hooks install
}

function copyToTemp() {
    local temp="$1"

    echo "Copy whole repo to temp and make one commit with all files."
    mkdir -p "$temp" &&
        cp -rf "$REPO_DIR" "$temp/repo" &&
        REPO_DIR_ORIG="$REPO_DIR" &&
        REPO_DIR="$temp/repo" &&
        cd "$REPO_DIR" &&
        rm -rf .git/hooks &&
        git commit --no-verify -a -m "Current working dir" &&
        git clean -fX &&
        rm -rf .git &&
        echo "Make repo..." &&
        git init &&
        echo "Make empty commit with tag" &&
        git commit --no-verify --allow-empty -m "Init" &&
        echo "Add all files." &&
        git add . &&
        git commit --no-verify -m "Original files" &&
        git tag v9.9.9
}

function generateAllFiles() {
    local src="$REPO_DIR/githooks"

    (cd "$src" && go mod vendor) || die "Go vendor failed."
    (cd "$src" && go generate -mod vendor ./...) || die "Could not generate."
}

function runAllHooks() {
    # Run all hooks.
    git checkout -b create-diffs &&
        git reset --soft HEAD~1 || die "Could not copy repo."
    git commit -m "Check all hooks." || die "Could not commit."
}

function fix() {
    echo "Copy diffing files back to fix them ..."
    readarray -t files < <(git --no-pager diff --name-only main)
    for file in "${files[@]}"; do
        echo "Copy file $file"
        cp "$file" "$REPO_DIR_ORIG/$file"
    done
}

function diff() {
    if [ "${GH_FIX:-false}" != "false" ]; then
        fix
    fi

    # Working tree diff to main .
    if ! git diff --quiet main; then
        [ "${GH_SHOW_DIFFS:-false}" == "false" ] || git --no-pager diff main

        die "Commit produced diffs, probably because of format" \
            "(use GH_SHOW_DIFFS=true to show diffs):?" \
            "$(git --no-pager diff --name-only main)"
    else
        echo "Checking all rules successful."
    fi
}

cleanUp

git config --global githooks.exportStagedFilesAsFile true
temp=$(mktemp -d)

copyToTemp "$temp"
installGithooks
generateAllFiles

deleteContainerVolumes
storeIntoContainerVolumes "$HOME/.githooks/shared"
setGithooksContainerVolumeEnvs "$temp/repo"
showAllContainerVolumes 2

runAllHooks

diff
