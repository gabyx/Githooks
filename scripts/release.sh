#!/usr/bin/env bash
set -euo pipefail

readarray -t prepareTag < <(git tag --pattern "prepare-*")

for tag in "${prepareTag[@]}"; do
    echo "Deleting prepare tag '$tag'."
    git push -f ":${tag}"
    git tag -d "$tag"
done

version="$1"
tag="v$version"
if git tag --pattern "v*" | grep -q "$tag"; then
    echo "Git tag '$tag' already exists."
fi

git tag "prepare-$tag"
git push -f origin "prepare-$tag"
