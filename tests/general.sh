#!/usr/bin/env bash

function die() {
    echo -e "! ERROR:" "$@" >&2
    exit 1
}

function accept_all_trust_prompts() {
    export ACCEPT_CHANGES=Y
    return 0
}

function is_docker_available() {
    command -v docker &>/dev/null || return 1
}

function is_image_existing() {
    docker image inspect "$1" &>/dev/null || return 1
}

function assert_no_test_images() {
    images=$(
        docker images -q --filter "reference=*test-image*" &&
            docker images -q --filter "reference=*/*test-image*" &&
            docker images -q --filter "reference=*/*/*test-image*"
    )

    if [ -n "$images" ]; then
        echo "Docker test images are still existing." >&2
        docker images --filter "reference=*test-image*" || true
        docker images --filter "reference=*/*test-image*" || true
        docker images --filter "reference=*/*/*test-image*" || true
        die "Failed."
    fi
}

function set_update_check_timestamp() {
    mkdir -p "$GH_INSTALL_DIR"
    echo "$1" >"$GH_INSTALL_DIR/.last-update-check-timestamp"
    return 0
}

function get_update_check_timestamp() {
    file="$GH_INSTALL_DIR/.last-update-check-timestamp"
    if [ -f "$file" ]; then
        cat "$file"
    fi

    return 0
}

function reset_update_check_timestamp() {
    rm "$GH_INSTALL_DIR/.last-update-check-timestamp" || return 0
}

function delete_all_test_images() {
    if ! is_docker_available; then
        return 0
    fi

    echo "Deleting all test images..."

    # Delete the images by the reference name, instead of ID,
    # because multiple tags to the same ID can exists.
    images=$(
        docker images -q --format="{{ .Repository }}:{{ .Tag }}" --filter "reference=*test-image*" &&
            docker images -q --format="{{ .Repository }}:{{ .Tag }}" --filter "reference=*/*test-image*" &&
            docker images -q --format="{{ .Repository }}:{{ .Tag }}" --filter "reference=*/*/*test-image*"
    )
    if [ -n "$images" ]; then
        # shellcheck disable=SC2086
        echo "$images" | while read -r img; do
            docker rmi -f "$img" >/dev/null || die "Could not delete images."
        done
    fi
}

function store_into_container_volumes() {
    local shared
    shared=$(cd "$1" && pwd)

    # copy content into volume
    store_into_container_volume "gh-test-shared" "$shared/."

    # Currently we do not need to store the workspace
    # because its in a volume from the start.
    # ==================================================
    # Copy the workspace into volume (under volume/repo)
    # we need this subfolder when entering
    # the container running in container `test-alpine-user`, if we would not have that
    # the directory (volume) would have root owner ship.
    # store_into_container_volume "gh-test-workspace" "$workspace" "./repo"
}

function store_into_container_volume() {
    local volume="$1"
    local src="$2" # Add a `/.` to copy the content.
    local dest="${3:-.}"
    echo "Storing '$src' into volume '$volume' for mounting."

    # shellcheck disable=SC2015
    docker volume create "$volume" &&
        docker container create --name githookscopytovolume \
            -v "$volume:/mnt/volume" githooks:volumecopy &&
        docker cp -a "$src" "githookscopytovolume:/mnt/volume/${dest}" &&
        docker container rm githookscopytovolume ||
        {
            docker container rm githookscopytovolume &>/dev/null || true
            die "Could not copy file to storage."
        }
}

function restore_from_container_volume() {
    local volume="$1"
    local base="$2"
    local dest="$3"
    shift 3
    local files=("$@")

    # shellcheck disable=SC2015
    docker container create --name githookscopytovolume \
        -v "$volume:/mnt/volume" githooks:volumecopy ||
        die "Could not start copy container."

    for file in "${files[@]}"; do
        echo "Restoring '$dest/$file' from volume path '$volume/$base/$file'."
        docker cp -a "githookscopytovolume:/mnt/volume/$base/$file" "$dest/$file" ||
            {
                docker container rm githookscopytovolume &>/dev/null || true
                die "Docker copy failed."
            }
    done

    docker container rm githookscopytovolume &>/dev/null ||
        die "Removing copycontainer failed."
}

function show_container_volume() {
    local volume="$1"
    local level="$2"

    # shellcheck disable=SC2015
    docker run --rm \
        -v "$volume:/mnt/volume" \
        -w "/mnt/volume" \
        alpine:latest \
        sh -c "apk add tree && echo Content of volume '$volume' && tree -aL $level" ||
        die "Could not show container volume."
}

function show_all_container_volumes() {
    local level="$1"
    show_container_volume "gh-test-shared" "$level"
}

function set_githooks_container_volume_envs() {
    local workspaceDest
    workspaceDest="$(cd "$1" && pwd)"

    local file
    file="$(mktemp)"

    cat <<<"
    workspace-path-dest: $workspaceDest
    # shared-path-dest: /mnt/shared # this is the default

    auto-mount-workspace: false
    auto-mount-shared: false

    args:
      - -v
      - gh-test-shared:/mnt/shared
      - -v
      - gh-test-tmp:/tmp
    " >"$file"

    export GITHOOKS_CONTAINER_RUN_CONFIG_FILE="$file"
}

function delete_container_volumes() {
    echo "Deleting all test container volumes ..."
    delete_container_volume "gh-test-shared"
}

function delete_container_volume() {
    local volume="$1"
    if docker volume ls | grep "$volume"; then
        docker volume rm "$volume"
    fi
}
