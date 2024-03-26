#!/usr/bin/env bash
set -euo pipefail

readarray -t prepareTag < <(git tag --list "prepare-*")

for tag in "${prepareTag[@]}"; do
    echo "Deleting prepare tag '$tag'."
    git push -f origin ":${tag}" || true
    git tag -d "$tag"
done

version="$1"
tag="v$version"
if git tag --list "v*" | grep -q "$tag"; then
    echo "Git tag '$tag' already exists."
fi

update_info="${1:-}"

if [ -z "$update_info" ]; then
    git tag "prepare-$tag"
else
    git tag -a -m "Version $tag" -m "Update-Info: $update_info" "prepare-$tag"
fi

echo "Pushing tag 'prepare-$tag'."
git push -f origin "prepare-$tag"
