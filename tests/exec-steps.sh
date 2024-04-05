#!/usr/bin/env bash
# shellcheck disable=SC1091

set -e
set -u

TEST_DIR=$(cd "$(dirname "$0")" && pwd)
. "$TEST_DIR/general.sh"

# shellcheck disable=SC2317
function clean_up() {
    set +e
    clean_dirs
    clean_docker
    set -e
}

function parse_args() {
    if [ "${1:-}" = "--skip-docker-check" ]; then
        shift
    else
        if [ "$DOCKER_RUNNING" != "true" ]; then
            echo "! This script is only meant to be run in a Docker container"
            exit 1
        fi
    fi

    if [ "${1:-}" = "--show-output" ]; then
        shift
        TEST_SHOW="true"
    fi

    if [ "${1:-}" = "--seq" ]; then
        shift
        SEQUENCE=$(for f in "$@"; do echo "step-$f"; done)
    fi
}

function clean_dirs() {
    if [ -d "$GH_TEST_GIT_CORE" ]; then
        # shellcheck disable=SC2015
        mkdir -p "$GH_TEST_GIT_CORE/templates/hooks" &&
            rm -rf "$GH_TEST_GIT_CORE/templates/hooks/"* ||
            {
                echo "! Cleanup failed."
                exit 1
            }
    fi

    # Delete all files in /tmp (/tmp might be a mount! cannot delete whole folder.)
    find /tmp -mindepth 1 -delete

    rm -rf ~/test* || true

    # Delete Githooks temp folder.
    if [ -d "$GH_TEST_TMP" ]; then
        rm -rf "$GH_TEST_TMP" || {
            echo "! Cleanup failed."
            exit 1
        }
    fi
    mkdir -p "$GH_TEST_TMP"

    return 0
}

function clean_docker() {
    if command -v "docker" &>/dev/null; then
        # shellcheck disable=SC2015
        delete_all_test_images &>/dev/null &&
            delete_container_volumes &>/dev/null || {
            echo "! Cleanup docker failed."
            exit 1
        }
    fi

    return 0
}

function reset_test_repo() {
    local commit_before="$1"

    # Reset test repo
    # shellcheck disable=SC2015
    git -C "$GH_TEST_REPO" -c core.hooksPath=/dev/null reset --hard "$commit_before" >/dev/null 2>&1 &&
        git -C "$GH_TEST_REPO" -c core.hooksPath=/dev/null clean -df || {
        echo "! Reset failed"
        exit 1
    }
}

function unset_environment() {
    # Unset mock settings
    git config --global --unset githooks.testingTreatFileProtocolAsRemote

    # Check if no githooks settings are present anymore
    if [ -n "$(git config --global --get-regexp "^githooks.*")" ] ||
        [ -n "$(git config --global alias.hooks)" ]; then
        echo "! Uninstall left artefacts behind!" >&2
        echo "  You need to fix this!" >&2
        git config --global --get-regexp "^githooks.*" >&2
        git config --global --get-regexp "alias.*" >&2
        return 1 # Fail es early as possible
    fi

    git config --global --unset init.templateDir
    git config --global --unset core.hooksPath
    rm -rf "$GH_INSTALL_DIR" 2>/dev/null || true

    return 0
}

function main() {

    local test_run=0
    local failed=0
    local skipped=0
    local failed_test_list=""

    if [ -z "${GH_TESTS:-}" ] ||
        [ -z "${GH_TEST_REPO:-}" ] ||
        [ -z "${GH_TEST_BIN:-}" ] ||
        [ -z "${GH_TEST_TMP:-}" ] ||
        [ -z "${GH_TEST_GIT_CORE:-}" ]; then
        echo "! Missing env. variables." >&2
        exit 1
    fi

    export GH_INSTALL_DIR="$HOME/.githooks"
    export GH_INSTALL_BIN_DIR="$GH_INSTALL_DIR/bin"

    local commit_before
    commit_before=$(git -C "$GH_TEST_REPO" rev-parse HEAD)

    echo "Test repo: '$GH_TEST_REPO'"
    echo "Tests dir: '$GH_TESTS'"
    echo "User: $(id -u)"
    echo "Group: $(id -g)"

    local startT endT
    startT=$(date +%s)

    for step in "$GH_TESTS/steps"/step-*.sh; do
        step_name=$(basename "$step" | sed 's/.sh$//')
        step_desc=$(grep -m 1 -A 1 "Test:" "$step" | tail -1 | sed 's/#\s*//')

        if [ -n "$SEQUENCE" ] && ! echo "$SEQUENCE" | grep -q "$step_name"; then
            continue
        fi

        echo "> Executing $step_name"
        echo "  :: $step_desc"

        clean_dirs
        clean_docker

        test_run=$((test_run + 1))

        {
            set +e
            test_output=$("$step" 2>&1)
            test_result=$?
            set -e
        }

        # shellcheck disable=SC2181
        if [ $test_result -eq 249 ]; then
            local reason
            reason=$(echo "$test_output" | tail -1)
            echo "  x  $step has been skipped, reason: $reason"
            skipped=$((skipped + 1))
        elif [ $test_result -eq 250 ]; then
            echo -e "  >  $step is benchmark:\n $test_output"
            skipped=$((skipped + 1))
        elif [ $test_result -ne 0 ]; then
            local failure
            failure=$(echo "$test_output" | tail -1)
            echo "! $step has failed with code $test_result ($failure), output:" >&2
            echo "$test_output" | sed -E "s/^/ x: /g" >&2
            failed=$((failed + 1))
            failed_test_list="$failed_test_list\n- $step ($test_result -- $failure)"

        elif [ "$TEST_SHOW" = "true" ]; then
            echo ":: Output was:"
            echo "$test_output" | sed -E "s/^/  | /g"
        fi

        if [ $test_result -eq 111 ]; then
            echo "! $step triggered fatal test abort." >&2
            break
        fi

        clean_dirs
        reset_test_repo "$commit_before"

        local uninstall_out
        uninstall_out=$(printf "n\\n" | "$GH_TEST_BIN/githooks-cli" uninstaller --stdin 2>&1)

        # shellcheck disable=SC2181
        if [ $? -ne 0 ]; then
            echo "! Uninstall failed in $step, output:" >&2
            echo "$uninstall_out" >&2
            failed=$((failed + 1))
            break # Fail es early as possible
        fi

        unset_environment || {
            echo -e "! Unset env. failed: uninstall output was:\n $uninstall_out" >&2
            failed=$((failed + 1))
            break
        }

        echo

    done

    endT=$(date +%s)
    local elapsed=$((endT - startT))

    if [ "$test_run" = "0" ]; then
        echo "No tests have been run which is a failure." >&2
        exit 1
    fi

    echo "$test_run tests run: $failed failed and $skipped skipped"
    echo "Run time: $elapsed seconds"
    echo

    if [ -n "$failed_test_list" ]; then
        echo -e "Failed tests: $failed_test_list" >&2
        echo
    fi

    if [ $failed -ne 0 ]; then
        exit 1
    else
        exit 0
    fi
}

trap clean_up EXIT

TEST_SHOW="false"
SEQUENCE=""
parse_args "$@"

main
