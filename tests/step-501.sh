#!/usr/bin/env bash
# Test:
#   Benchmark runner with no load

TEST_DIR=$(cd "$(dirname "$0")" && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

if [ -n "$GH_COVERAGE_DIR" ]; then
    echo "Benchmark not for coverage."
    exit 249
fi

acceptAllTrustPrompts || exit 1

git -C "$GH_TEST_REPO" reset --hard v2.1.0 >/dev/null 2>&1 || exit 1

# run the default install
"$GH_TEST_BIN/cli" installer &>/dev/null || exit 1

# Overwrite runner.
git config --global "githooks.runner" "$GH_TEST_BIN/runner"

mkdir -p "$GH_TEST_TMP/test501" &&
    cd "$GH_TEST_TMP/test501" &&
    git init || exit 1

function runCommits() {
    for i in {1..30}; do
        git commit --allow-empty -m "Test $i" 2>&1 | average
    done
}

function calc() { awk "BEGIN{print $*}"; }

function average() {
    local count=0
    local total=0

    local input
    # Skip the first 3 runs, because it contains registration etc...
    input=$(cat | grep "execution time:" | sed -E "s/.*'(.*)ms.*/\1/g" | sed -n '3d;p')

    while IFS= read -r val; do
        total=$(calc "$total" + "$val")
        ((count++))
    done <<<"$input"

    time=$(calc "$total" / "$count")

    echo "execution time: '$time""ms'"
}

echo "Runtime average (no load): $(runCommits | average)"

exit 250
