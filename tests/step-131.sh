#!/usr/bin/env bash
# Test:
#   Test .envs.yaml with shared hooks.

TEST_DIR=$(cd "$(dirname "$0")" && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

acceptAllTrustPrompts || exit 1

git config --global githooks.testingTreatFileProtocolAsRemote "true"

# Hooks with applied envs.yaml
# shellcheck disable=SC2016
mkdir -p "$GH_TEST_TMP/shared/hooks-131a.git/githooks/pre-commit" &&
    cd "$GH_TEST_TMP/shared/hooks-131a.git" &&
    echo 'env > "$GH_TEST_TMP/envs-a" && exit 0' >"githooks/pre-commit/print-envs" &&
    echo 'mystuff-a' >"githooks/.namespace" &&
    cd "$GH_TEST_TMP/shared/hooks-131a.git" &&
    git init && git add . && git commit -m 'Initial commit' ||
    exit 1

# Hooks without applied envs.
# shellcheck disable=SC2016
mkdir -p "$GH_TEST_TMP/shared/hooks-131b.git/githooks/pre-commit" &&
    cd "$GH_TEST_TMP/shared/hooks-131b.git" &&
    echo 'env > "$GH_TEST_TMP/envs-b" && exit 0' >"githooks/pre-commit/print-envs" &&
    echo 'mystuff-b' >"githooks/.namespace" &&
    cd "$GH_TEST_TMP/shared/hooks-131b.git" &&
    git init && git add . && git commit -m 'Initial commit' ||
    exit 1

mkdir -p "$GH_TEST_TMP/test18" &&
    cd "$GH_TEST_TMP/test18" &&
    git init || exit 1

mkdir -p .githooks || exit 1
cat <<EOF >.githooks/.shared.yaml || exit 1
urls:
    - file://$GH_TEST_TMP/shared/hooks-131a.git
    - file://$GH_TEST_TMP/shared/hooks-131b.git
EOF

cat <<EOF >.githooks/.envs.yaml || exit 1
envs:
    mystuff-a:
        - MYSTUFF_A1=aaa
        - MYSTUFF_A2=bbb
    mystuff-b:
        - MYSTUFF_B1=ccc
        - MYSTUFF_B2=ddd
EOF

"$GH_TEST_BIN/cli" shared update || exit 1

# Execute pre-commit and check that env variables are applied.
"$GH_TEST_BIN/runner" "$(pwd)"/.git/hooks/pre-commit || exit 1

# shellcheck disable=SC2015
grep -q "MYSTUFF_A1=aaa" "$GH_TEST_TMP/envs-a" &&
    grep -q "MYSTUFF_A2=bbb" "$GH_TEST_TMP/envs-a" &&
    ! grep "MYSTUFF_B" "$GH_TEST_TMP/envs-a" ||
    {
        echo "Wrong env variables:"
        cat "$GH_TEST_TMP/envs-a"
        exit 1
    }

# shellcheck disable=SC2015
grep "MYSTUFF_B1=ccc" "$GH_TEST_TMP/envs-b" &&
    grep "MYSTUFF_B2=ddd" "$GH_TEST_TMP/envs-b" &&
    ! grep "MYSTUFF_A" "$GH_TEST_TMP/envs-b" ||
    {
        echo "Wrong env variables:"
        cat "$GH_TEST_TMP/envs-b"
        exit 1
    }
