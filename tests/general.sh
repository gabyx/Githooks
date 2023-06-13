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

function storeIntoContainerVolume() {
    volume="$1"
    src="$2/." # Add a `/.` to copy the content.

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

function deleteContainerVolumes() {
    docker volume rm "gh-test-workspace" &>/dev/null || true
    docker volume rm "gh-test-shared" &>/dev/null || true
    return 0
}
