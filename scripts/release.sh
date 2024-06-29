#!/usr/bin/env bash
#
# Creating a prepare tag to trigger the release process on the
# Github workflow.
# You can set the following meaning full
# Git trailers on the annotated tag:
# `Update-Info: ...` : One line of update info to be given to the user when updating.
# `Update-NoSkip: true` : Specifying that this version update cannot be skipped.
# `Release-Branch: .*`: Specifying another branch than main.
#                       Useful for testing stuff which does not affect the user, since
#                       it must be on this branch to get this update.
#                       This branch will be checked to contain the tag.

set -euo pipefail

ROOT_DIR=$(git rev-parse --show-toplevel)
cd "$ROOT_DIR"

function delete_prepare_tags() {
    readarray -t prepareTag < <(git tag --list "prepare-*")

    for tag in "${prepareTag[@]}"; do
        echo "Deleting prepare tag '$tag'."
        git push -f origin ":${tag}" || true
        git tag -d "$tag"
    done
}

function commit_version_file() {
    local version="$1"
    echo "Writing new version file..."

    temp=$(mktemp)
    jq ".version |= \"$version\"" nix/pkgs/version.json >"$temp"
    mv "$temp" nix/pkgs/version.json

    if ! git diff --quiet --exit-code; then
        git add nix/pkgs/version.json
        git commit -m "np: Update version to '$version'"
    fi
}

function create_tag() {
    tag="v$version"
    if git tag --list "v*" | grep -qE "^$tag$"; then
        echo "Git tag '$tag' already exists."
        exit 1
    fi

    if git ls-remote "refs/tags/v*" | grep -qE "^$tag$"; then
        echo "Git tag '$tag' already exists."
        exit 1
    fi

    add_message=()
    if [ -n "$update_info" ]; then
        add_message+=(-m "Update-Info: $update_info")
    fi

    if [ "$branch" != "main" ]; then
        add_message+=(-m "Release-Branch: $branch")
    fi

    echo "Tagging..."
    git tag -a -m "Version $tag" -m "${add_message[@]}" "prepare-$tag"

    echo "Tag contains:"
    git cat-file -p "prepare-$tag"
}

function trigger_build() {
    printf "Do you want to trigger the build? [y|n]: "
    read -r answer
    if [ "$answer" != "y" ]; then
        echo "Do not trigger build -> abort."
        exit 0
    fi

    echo "Pushing tag 'prepare-$tag'."
    git push -f origin --no-follow-tags "$branch" "prepare-$tag"
}

version="$1"
update_info="${2:-}"
branch=$(git branch --show-current)

if ! git diff --quiet --exit-code; then
    echo "You have changes on this branch."
    exit 1
fi

delete_prepare_tags
commit_version_file "$version"
create_tag
trigger_build
