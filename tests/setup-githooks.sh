#!/usr/bin/env bash

set -e
set -u

cliAddArgs=()
buildAddArgs=()
if [ "${1:-}" = "--coverage" ]; then
    buildAddArgs+=("--coverage")
    cliAddArgs+=("githooksCoverage")
fi

git config --global user.email "githook@test.com" &&
    git config --global user.name "Githook Tests" &&
    git config --global init.defaultBranch main &&
    git config --global core.autocrlf false || exit 1

rm -rf "$GH_TEST_REPO/.git" || true
echo "Make test Git repo to clone from ..." &&
    # We use the bin folder
    cd "$GH_TEST_REPO" &&
    sed -i -E 's/^\*//' "githooks/bin/.gitignore" &&
    git init >/dev/null 2>&1 &&
    git add . >/dev/null 2>&1 &&
    git commit -a -m "Before build" >/dev/null 2>&1 || exit 1

# Make a build which exists on the server on branch "test-download"
cd "$GH_TEST_REPO/githooks" &&
    git checkout -b "test-download" &&
    git commit -a --allow-empty \
        -m "Build version 2.0.0 for download test over Github" >/dev/null 2>&1 &&
    git tag "v2.0.0" &&
    ./scripts/clean.sh &&
    ./scripts/build.sh "${buildAddArgs[@]}" --prod &&
    ./bin/cli "${cliAddArgs[@]}" --version || exit 1
echo "Commit build v2.0.0 to repo (for test download) ..." &&
    cd "$GH_TEST_REPO" &&
    git add . >/dev/null 2>&1 &&
    git commit -a --allow-empty -m "Version 2.0.0" >/dev/null 2>&1 &&
    git tag -f "v2.0.0" || exit 1

# Setup server repository from which we install updates
# branch: main
# Build binaries for v9.9.0.
#################################
cd "$GH_TEST_REPO/githooks" &&
    git checkout main &&
    git tag "v9.9.0" &&
    ./scripts/clean.sh &&
    ./scripts/build.sh "${buildAddArgs[@]}" --build-tags "mock" &&
    ./bin/cli $"${cliAddArgs[@]}" --version || exit 1
echo "Commit build v9.9.0 to repo ..." &&
    cd "$GH_TEST_REPO" &&
    git add . >/dev/null 2>&1 &&
    git commit -a --allow-empty -m "Version 9.9.0" >/dev/null 2>&1 &&
    git tag -f "v9.9.0" || exit 1

# Build binaries for v9.9.1.
#################################
cd "$GH_TEST_REPO/githooks" &&
    git commit -a --allow-empty -m "Before build" >/dev/null 2>&1 &&
    git tag -f "v9.9.1" &&
    ./scripts/clean.sh &&
    ./scripts/build.sh "${buildAddArgs[@]}" --build-tags "mock" &&
    ./bin/cli "${cliAddArgs[@]}" --version || exit 1
echo "Commit build v9.9.1 to repo (no-skip)..." &&
    cd "$GH_TEST_REPO" &&
    git commit -a --allow-empty -m "Version 9.9.1" -m "Update-NoSkip: true" >/dev/null 2>&1 &&
    git tag -f "v9.9.1" || exit 1

# Commit for to v9.9.2 (not used for update).
#################################
echo "Commit build v9.9.2 to repo ..." &&
    cd "$GH_TEST_REPO" &&
    git commit -a --allow-empty -m "Version 9.9.2" \
        -m "Update-Info: Bug fixes and improvements." >/dev/null 2>&1 &&
    git tag -f "v9.9.2"

# Commit for to v10.1.1 (build not used).
#################################
echo "Commit build v10.1.1 to repo ..." &&
    cd "$GH_TEST_REPO" &&
    git commit -a --allow-empty -m "Version v10.1.1" \
        -m "Update-Info: Breaking changes, read the change log." >/dev/null 2>&1 &&
    git tag -f "v10.1.1" || exit 1

echo "Reset main to 9.9.0 ..." &&
    git reset --hard v9.9.0 || exit 1
