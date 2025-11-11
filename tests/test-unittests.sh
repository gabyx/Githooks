#!/usr/bin/env bash
# shellcheck disable=SC1091
#
set -e
set -u

ROOT_DIR=$(git rev-parse --show-toplevel)
. "$ROOT_DIR/tests/general.sh"

cd "$ROOT_DIR"

function clean_up() {
    # shellcheck disable=SC2317
    docker rmi "githooks:unittests" &>/dev/null || true
}

trap clean_up EXIT

cd "$ROOT_DIR"

cat <<EOF | docker build --force-rm -t githooks:unittests -
FROM golang:1.24-alpine
RUN apk update && apk add git git-lfs
RUN apk add bash jq curl docker

# CVE https://github.blog/2022-10-18-git-security-vulnerabilities-announced/#cve-2022-39253
RUN git config --system protocol.file.allow always


# Git refuses to do stuff in this mounted directory.
RUN git config --global safe.directory /githooks

ENV DOCKER_RUNNING=true
EOF

if ! docker run --rm -it \
    -v "$(pwd)":/githooks \
    -v "/var/run/docker.sock:/var/run/docker.sock" \
    -w /githooks githooks:unittests \
    tests/exec-unittests.sh; then

    echo "! Check rules had failures."
    exit 1
fi

exit 0
