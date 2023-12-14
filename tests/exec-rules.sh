#!/usr/bin/env bash
# shellcheck disable=SC2015

set -e
set -u

true && false && true

if [ "${DOCKER_RUNNING:-}" != "true" ]; then
    echo "! This script is only meant to be run in a Docker container"
    exit 1
fi

DIR="$(cd "$(dirname "$0")" >/dev/null 2>&1 && pwd)"
REPO_DIR="$DIR/.."

# shellcheck disable=SC1091
. "$DIR/general.sh"

# shellcheck disable=SC2317
function cleanUp() {
    set +e
    deleteContainerVolumes
}

[ "${GH_SHOW_DIFFS:-false}" == "false" ] || echo "INFO: SHOWING DIFFS"

trap cleanUp EXIT

echo "Copy whole repo to temp and make one commit with all files."
temp=$(mktemp -d)
cp -rf "$REPO_DIR" "$temp/repo" &&
    REPO_DIR="$temp/repo" &&
    cd "$REPO_DIR" &&
    rm -rf .git &&
    echo "Make repo..." &&
    git init &&
    echo "Make empty commit with tag" &&
    GITHOOKS_DISABLE=1 git commit --no-verify --allow-empty -m "Init" &&
    git hooks config trust-all --accept &&
    git hooks config enable-containerized-hooks --set &&
    git hooks shared update &&
    echo "Add all files." &&
    git add . &&
    GITHOOKS_DISABLE=1 git commit --no-verify -m "Original files" &&
    git checkout -b create-diffs &&
    git reset --soft HEAD~1 || die "Could not copy repo"

function setupGo() {
    local src="$REPO_DIR/githooks"

    (cd "$src" && go mod vendor) || die "Go vendor failed."

    GITHOOKS_DISABLE=1 git tag v9.9.9 &&
        (cd "$src" && go generate -mod vendor ./...) || die "Could not generate."
}

cd "$REPO_DIR" || exit 1

setupGo

deleteContainerVolumes
storeIntoContainerVolumes "$REPO_DIR" "$HOME/.githooks/shared" # for dockerized containers
setGithooksContainerVolumeEnvs

git commit -m "Check all hooks."

restoreFromContainerVolumeWorkspace "." ""

if ! git diff --quiet main..create-diffs; then
    [ "${GH_SHOW_DIFFS:-false}" == "false" ] || git diff main..create-diffs

    die "Commit produced diffs, probably because of format" \
        "(use GH_SHOW_DIFFS=true to show diffs):" \
        "$(git diff --name-only main..create-diffs)"
fi

if ! git diff --cached --quiet main; then
    [ "${GH_SHOW_DIFFS:-false}" == "false" ] || git diff --cached --name-only main

    die "Commit produced diffs, probably because of format" \
        "(use GH_SHOW_DIFFS=true to show diffs):?" \
        "$(git diff --cached --name-only main)"
fi

deleteContainerVolumes
exit 0
