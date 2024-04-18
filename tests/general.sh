#!/usr/bin/env bash

function init_step() {
    # Set extra install arguments for all steps.
    # when running centralized tests.
    if [ "${GH_TEST_CENTRALIZED_INSTALL:-}" = "true" ]; then
        echo "Setting extra install args to '--centralized'"
        # shellcheck disable=SC2034
        EXTRA_INSTALL_ARGS=(--centralized)
    fi
}

function is_centralized_tests() {
    [ "${GH_TEST_CENTRALIZED_INSTALL:-}" = "true" ] || return 1
}

function die() {
    echo -e "! ERROR:" "$@" >&2
    exit 1
}

function check_paths_are_equal() {
    local a
    a=$(echo "$1" | wrap_windows_paths)
    local b
    b=$(echo "$1" | wrap_windows_paths)
    shift 2

    [ "$a" = "$b" ] || {
        echo "! Paths '$a' != '$b' :" "$@"
        exit 1
    }
}

function wrap_windows_paths() {
    # On Windows we need to wrap paths sometimes.
    # to make them equivalent to the Git bash thing.
    cat | sed -E "s@[Cc]:/@/c/@g"
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
    local workspace_dest
    workspace_dest="$(cd "$1" && pwd)"

    local file
    file="$(mktemp)"

    cat <<<"
    workspace-path-dest: $workspace_dest
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

# Run the container manager which is defined.
function container_mgr() {
    if command -v podman &>/dev/null; then
        podman "$@"
    else
        docker "$@"
    fi

}

function install_hooks_if_not_centralized() {
    if ! is_centralized_tests; then
        "$GH_INSTALL_BIN_DIR/githooks-cli" install "$@" || {
            echo "! Could not install hooks (not centralized)."
            exit 1
        }
    fi
}

function check_normal_install {
    check_install "$@"

    if [ -n "$(git config --global core.hooksPath)" ]; then
        echo "! Global config 'core.hooksPath' must not be set."
        git config --global core.hooksPath
        exit 1
    fi
}

function check_centralized_install() {
    local path_to_use
    local expected
    expected="${2:-}"

    path_to_use=$(git config --global githooks.pathForUseCoreHooksPath)
    [ -d "$path_to_use" ] || {
        echo "! Path '$path_to_use' does not exist."
        exit 1
    }

    check_paths_are_equal "$path_to_use" "$(git config --global core.hooksPath)" \
        "! Global config 'core.hooksPath' does not point to the same directory."

    if [ -n "$expected" ] && [ "$path_to_use" != "$expected" ]; then
        echo "! Path '$path_to_use' is not '$expected'"
        exit 1
    fi

}

function check_no_local_install() {
    local dir
    local repo="$1"

    dir=$(git -C "$repo" rev-parse --path-format=absolute --git-common-dir) || {
        echo "! Failed to get Git dir."
        exit 1
    }

    if [ ! -d "$dir" ]; then
        echo "! Dir '$dir' does not exist."
        exit 1
    fi

    if grep -rq 'github.com/gabyx/githooks' "$dir"; then
        echo "! Githooks were installed into '$dir' but should not have:"
        grep -r 'github.com/gabyx/githooks' "$dir"
        exit 1
    fi

    if [ -f "$dir/githooks-contains-run-wrappers" ]; then
        echo "! Run-wrapper marker file is existing in '%dir'."
        exit 1
    fi

    if [ -n "$(git -C "$dir" config --local core.hooksPath)" ]; then
        echo "! Local config 'core.hooksPath' is set but should not:" \
            "'$(git -C "$dir" config --local core.hooksPath)'"
        exit 1
    fi
}

function check_local_install() {
    local repo="${1:-.}"
    local expected="${2:-}"

    dir=$(git -C "$repo" rev-parse --path-format=absolute --git-common-dir) || {
        echo "! Failed to get Git dir."
        exit 1
    }

    if [ ! -d "$dir" ]; then
        echo "! Dir '$dir' does not exist."
        exit 1
    fi

    local path_to_use
    path_to_use=$(git config --global githooks.pathForUseCoreHooksPath)
    [ -d "$path_to_use" ] || {
        echo "! Path '$path_to_use' does not exist."
        exit 1
    }

    check_paths_are_equal "$path_to_use" "$(git -C "$dir" config --local core.hooksPath)" \
        "Local config 'core.hooksPath' in '$repo' does not point to same directory."

    if [ -n "$expected" ] && [ "$path_to_use" != "$expected" ]; then
        echo "! Path '$path_to_use' is not '$expected'"
        exit 1
    fi
}

function check_local_install_no_run_wrappers() {
    local repo="${1:-.}"

    dir=$(git -C "$repo" rev-parse --path-format=absolute --git-common-dir) || {
        echo "! Failed to get Git dir."
        exit 1
    }

    if [ ! -d "$dir" ]; then
        echo "Dir '$dir' does not exist."
        exit 1
    fi

    if grep -rq 'github.com/gabyx/githooks' "$dir"; then
        echo "! Githooks were installed into '$dir'."
        exit 1
    fi

    if [ -z "$(git -C "$dir" config --local core.hooksPath)" ]; then
        echo "! Local config 'core.hooksPath' in '$repo' should be set."
        git -C "$dir" config --local core.hooksPath
        exit 1
    fi
}

function check_local_install_run_wrappers() {
    local repo="${1:-.}"

    dir=$(git -C "$repo" rev-parse --path-format=absolute --git-common-dir) || {
        echo "! Failed to get Git dir."
        exit 1
    }

    if [ ! -d "$dir" ]; then
        echo "! Dir '$dir' does not exist."
        exit 1
    fi

    if ! grep -rq 'github.com/gabyx/githooks' "$dir"; then
        echo "! Githooks were not installed into '$dir'."
        exit 1
    fi

    if [ -n "$(git -C "$dir" config --local core.hooksPath)" ]; then
        echo "! Config 'core.hooksPath' in '$repo' should not be set."
        git -C "$dir" config --local core.hooksPath
        exit 1
    fi
}
function check_no_install() {
    if git config --get-regexp "^githooks.*" ||
        [ -n "$(git config --global alias.hooks)" ]; then

        echo "Should not have set Git config variables."
        git config --get-regexp "^githooks.*"

        exit 1
    fi
}

function check_install() {
    local expected
    expected="${1:-}"

    local path_to_use
    path_to_use=$(git config --global githooks.pathForUseCoreHooksPath)
    [ -d "$path_to_use" ] || {
        echo "! Path '$path_to_use' does not exist."
        exit 1
    }

    if ! grep -rq 'github.com/gabyx/githooks' "$path_to_use"; then
        echo "! Githooks were not installed into '$path_to_use'."
        exit 1
    fi

    if [ -n "$expected" ]; then
        check_paths_are_equal "$path_to_use" "$expected"
    fi

    if [ ! -f "$path_to_use/githooks-contains-run-wrappers" ]; then
        echo "! Folder '$path_to_use' should contain a marker file."
        ls -al "$path_to_use"
        exit 1
    fi
}

function check_install_hooks_local() {
    local repo="$1"
    local count_expected="$2"
    shift 2
    local hook_names=("$@")

    dir=$(git -C "$repo" rev-parse --path-format=absolute --git-common-dir) || {
        echo "! Failed to get Git dir."
        exit 1
    }

    if [ ! -d "$dir" ]; then
        echo "! Dir '$dir' does not exist."
        exit 1
    fi

    local path_to_use="$dir/hooks"

    for hook in "${hook_names[@]}"; do
        if [ ! -f "$path_to_use/$hook" ]; then
            echo "! Hooks '$hook' was not installed successfully in '$path_to_use'."
            exit 1
        fi
    done

    # shellcheck disable=SC2012
    count=$(find "$path_to_use" -type f -not -name "githooks-contains-run-wrappers" | wc -l)
    if [ "$count" != "$count_expected" ]; then
        echo "! Expected only '$count_expected' to be installed ($count)"
        find "$path_to_use" -type f -not -name "githooks-contains-run-wrappers"
        exit 1
    fi
}

function check_install_hooks() {
    local count_expected="$1"
    shift 1
    local hook_names=("$@")

    local path_to_use
    path_to_use=$(git config --global githooks.pathForUseCoreHooksPath)
    [ -d "$path_to_use" ] || {
        echo "! Path '$path_to_use' does not exist."
        exit 1
    }

    for hook in "${hook_names[@]}"; do
        if [ ! -f "$path_to_use/$hook" ]; then
            echo "! Hooks '$hook' was not installed successfully in '$path_to_use'."
            exit 1
        fi
    done

    # shellcheck disable=SC2012
    count=$(find "$path_to_use" -type f -not -name "githooks-contains-run-wrappers" | wc -l)
    if [ "$count" != "$count_expected" ]; then
        echo "! Expected only '$count_expected' to be installed ($count)"
        find "$path_to_use" -type f -not -name "githooks-contains-run-wrappers"
        exit 1
    fi
}
