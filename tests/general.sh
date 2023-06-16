#!/usr/bin/env bash

function acceptAllTrustPrompts() {
    export ACCEPT_CHANGES=Y
    return 0
}

function isDockerAvailable() {
    command -v docker &>/dev/null || return 1
}

function isImageExisting() {
    docker image inspect "$1" &>/dev/null || return 1
}

function assertNoTestImages() {
    images=$(
        docker images -q --filter "reference=*test-image*" &&
            docker images -q --filter "reference=*/*test-image*"
        docker images -q --filter "reference=*/*/*test-image*"
    )

    if [ -n "$images" ]; then
        echo "Docker test images are still existing." >&2
        docker images --filter "reference=*test-image*"
        docker images --filter "reference=*/*test-image*"
        docker images --filter "reference=*/*/*test-image*"
        exit 1
    fi
}

function deleteAllTestImages() {
    if ! isDockerAvailable; then
        return 0
    fi

    # Delete the images by the reference name, instead of ID,
    # because multiple tags to the same ID can exists.
    images=$(
        docker images -q --format="{{ .Repository }}:{{ .Tag }}" --filter "reference=*test-image*" &&
            docker images -q --format="{{ .Repository }}:{{ .Tag }}" --filter "reference=*/*test-image*"
        docker images -q --format="{{ .Repository }}:{{ .Tag }}" --filter "reference=*/*/*test-image*"
    )
    if [ -n "$images" ]; then
        # shellcheck disable=SC2086
        echo "$images" | while read -r img; do
            docker rmi -f "$img" >/dev/null
        done
    fi
}

function storeIntoContainerVolumes() {
    local workspace
    workspace=$(cd "$1" && pwd)
    local shared
    shared=$(cd "$2" && pwd)

    storeIntoContainerVolume "gh-test-workspace" "$workspace" # copy folder into volume (not content)
    storeIntoContainerVolume "gh-test-shared" "$shared"       # copy folder into volume
}

function restoreFromContainerVolumeWorkspace() {
    local workspace="$1"
    local file="$2"
    restoreFromContainerVolume "gh-test-workspace" "$workspace" "$file"
}

function storeIntoContainerVolume() {
    volume="$1"
    src="$2/" # Add a `/.` to copy the content.
    echo "Storing '$src' into volume '$volume' for mounting."

    # shellcheck disable=SC2015
    docker volume create "$volume" &&
        docker run -d --rm --name githookscopytovolume \
            -v "$volume:/mnt/volume" alpine:latest tail -f /dev/null &&
        docker cp -a "$src" "githookscopytovolume:/mnt/volume" &&
        docker stop githookscopytovolume ||
        {
            docker stop githookscopytovolume &>/dev/null || true
            return 1
        }
    return 0
}
function restoreFromContainerVolume() {
    volume="$1"
    dest="$2"
    file="$3"

    echo "Restoring '$dest/$file' from volume '$volume'."

    # shellcheck disable=SC2015
    docker run -d --rm --name githookscopytovolume \
        -v "$volume:/mnt/volume" alpine:latest tail -f /dev/null &&
        docker cp -a "githookscopytovolume:/mnt/volume/$file" "$dest/$file" &&
        docker stop githookscopytovolume ||
        {
            docker stop githookscopytovolume &>/dev/null || true
            return 1
        }
    return 0
}

function setGithooksContainerVolumeEnvs() {
    # Use a volume for the host path.
    export GITHOOKS_CONTAINER_WORKSPACE_HOST_PATH="gh-test-workspace"
    export GITHOOKS_CONTAINER_WORKSPACE_BASE_PATH="./\${repository-dir-name}"
    export GITHOOKS_CONTAINER_SHARED_HOST_PATH="gh-test-shared"
}

function deleteContainerVolumes() {
    docker volume rm "gh-test-workspace" &>/dev/null || true
    docker volume rm "gh-test-shared" &>/dev/null || true
    return 0
}
