#!/usr/bin/env bash

if [ "$1" = "--skip-docker-check" ]; then
    shift
else
    if ! grep '/docker/' </proc/self/cgroup >/dev/null 2>&1; then
        echo "! This script is only meant to be run in a Docker container" >&2
        exit 1
    fi
fi

if [ "$1" = "--show-output" ]; then
    shift
    TEST_SHOW="true"
fi

SEQUENCE=""
if [ "$1" = "--seq" ]; then
    shift
    SEQUENCE=$(for f in "$@"; do echo "step-$f"; done)
fi

TEST_RUNS=0
FAILED=0
SKIPPED=0

FAILED_TEST_LIST=""

export GH_INSTALL_DIR="$HOME/.githooks"
export GH_INSTALL_BIN_DIR="$GH_INSTALL_DIR/bin"
COMMIT_BEFORE=$(git -C "$GH_TEST_REPO" rev-parse HEAD)

function cleanDirs() {

    if [ -d "$GH_TEST_GIT_CORE" ]; then
        mkdir -p "$GH_TEST_GIT_CORE/templates/hooks" || {
            echo "! Cleanup failed."
            exit 1
        }
        rm -rf "$GH_TEST_GIT_CORE/templates/hooks/"*
    fi

    rm -rf /tmp/githooks-installer-*
    rm -rf ~/test*
    rm -rf "$GH_TEST_TMP"
    mkdir -p "$GH_TEST_TMP" || {
        echo "! Cleanup failed."
        exit 1
    }
}

function resetTestRepo() {
    # Reset test repo
    # shellcheck disable=SC2015
    git -C "$GH_TEST_REPO" -c core.hooksPath=/dev/null reset --hard "$COMMIT_BEFORE" >/dev/null 2>&1 &&
        git -C "$GH_TEST_REPO" -c core.hooksPath=/dev/null clean -df || {
        echo "! Reset failed"
        exit 1
    }
}

function unsetEnvironment() {
    # Unset mock settings
    git config --global --unset githooks.testingTreatFileProtocolAsRemote

    # Check if no githooks settings are present anymore
    if [ -n "$(git config --global --get-regexp "^githooks.*")" ] ||
        [ -n "$(git config --global alias.hooks)" ]; then
        echo "! Uninstall left artefacts behind!" >&2
        echo "  You need to fix this!" >&2
        git config --global --get-regexp "^githooks.*" >&2
        git config --global --get-regexp "alias.*" >&2
        FAILED=$((FAILED + 1))
        return 1 # Fail es early as possible
    fi

    git config --global --unset init.templateDir
    git config --global --unset core.hooksPath
    rm -rf "$GH_INSTALL_DIR" 2>/dev/null || true

    return 0
}

if [ -z "$GH_TESTS" ] ||
    [ -z "$GH_TEST_REPO" ] ||
    [ -z "$GH_TEST_BIN" ] ||
    [ -z "$GH_TEST_TMP" ] ||
    [ -z "$GH_TEST_GIT_CORE" ]; then
    echo "! Missing env. variables." >&2
    exit 1
fi

echo "Test repo: '$GH_TEST_REPO'"
echo "Tests dir: '$GH_TESTS'"

startT=$(date +%s)

for STEP in "$GH_TESTS"/step-*.sh; do
    STEP_NAME=$(basename "$STEP" | sed 's/.sh$//')
    STEP_DESC=$(grep -m 1 -A 1 "Test:" "$STEP" | tail -1 | sed 's/#\s*//')

    if [ -n "$SEQUENCE" ] && ! echo "$SEQUENCE" | grep -q "$STEP_NAME"; then
        continue
    fi

    echo "> Executing $STEP_NAME"
    echo "  :: $STEP_DESC"

    cleanDirs

    TEST_RUNS=$((TEST_RUNS + 1))

    TEST_OUTPUT=$("$STEP" 2>&1)
    TEST_RESULT=$?
    # shellcheck disable=SC2181
    if [ $TEST_RESULT -eq 249 ]; then
        REASON=$(echo "$TEST_OUTPUT" | tail -1)
        echo "  x  $STEP has been skipped, reason: $REASON"
        SKIPPED=$((SKIPPED + 1))

    elif [ $TEST_RESULT -ne 0 ]; then
        FAILURE=$(echo "$TEST_OUTPUT" | tail -1)
        echo "! $STEP has failed with code $TEST_RESULT ($FAILURE), output:" >&2
        echo "$TEST_OUTPUT" | sed -E "s/^/ x: /g" >&2
        FAILED=$((FAILED + 1))
        FAILED_TEST_LIST="$FAILED_TEST_LIST
- $STEP ($TEST_RESULT -- $FAILURE)"

    elif [ -n "$TEST_SHOW" ]; then
        echo ":: Output was:"
        echo "$TEST_OUTPUT" | sed -E "s/^/  | /g"
    fi

    if [ $TEST_RESULT -eq 111 ]; then
        echo "! $STEP triggered fatal test abort." >&2
        break
    fi

    cleanDirs
    resetTestRepo

    UNINSTALL_OUTPUT=$(printf "n\\n" | "$GH_TEST_BIN/cli" uninstaller --stdin 2>&1)
    # shellcheck disable=SC2181
    if [ $? -ne 0 ]; then
        echo "! Uninstall failed in $STEP, output:" >&2
        echo "$UNINSTALL_OUTPUT" >&2
        FAILED=$((FAILED + 1))
        break # Fail es early as possible
    fi

    unsetEnvironment || break

    echo

done

endT=$(date +%s)
elapsed=$((endT - startT))

echo "$TEST_RUNS tests run: $FAILED failed and $SKIPPED skipped"
echo "Run time: $elapsed seconds"
echo

if [ -n "$FAILED_TEST_LIST" ]; then
    echo "Failed tests: $FAILED_TEST_LIST" >&2
    echo
fi

if [ $FAILED -ne 0 ]; then
    exit 1
else
    exit 0
fi
