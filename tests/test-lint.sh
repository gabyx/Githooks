#!/usr/bin/env bash
# shellcheck disable=SC1091

set -e
set -u

ROOT_DIR=$(git rev-parse --show-toplevel)
. "$ROOT_DIR/tests/general.sh"

cd "$ROOT_DIR"

# shellcheck disable=SC2317
function clean_up() {
    run_docker rmi "githooks:test-rules" &>/dev/null || true
    run_docker volume rm gh-test-tmp &>/dev/null || true
}

trap clean_up EXIT

clean_up

# Build container to only copy to volumes.
cat <<EOF | run_docker build \
    --force-rm -t "githooks:volumecopy" -f - . || exit 1
    FROM scratch
    CMD you-should-not-run-this-container
EOF

# Build test container.
cat <<EOF | run_docker build --force-rm -t githooks:test-rules -f - .
FROM alpine:3.23.0
RUN apk update && apk add curl git

RUN git config --global safe.directory /data

# Install Nix
RUN curl --proto '=https' --tlsv1.2 -sSf -L https://install.determinate.systems/nix | sh -s -- install linux \
  --extra-conf "sandbox = false" \
  --init none \
  --no-confirm
ENV PATH="/nix/var/nix/profiles/default/bin:\$PATH"
RUN nix --version

ADD flake.nix flake.lock /
RUN nix --accept-flake-config develop ".#default" --command true

ENV DOCKER_RUNNING=true
EOF

# Create a volume where all test setup and repositories go in.
# Is mounted to `/tmp`
run_docker volume create gh-test-tmp

# Always show diffs.
export GH_SHOW_DIFFS=true

mountArg=":ro"
if [ "${GH_FIX:-}" = "true" ]; then
    mountArg=""
fi

run_docker run --rm -it \
    -v "$ROOT_DIR:/data$mountArg" \
    -v "gh-test-tmp:/tmp" \
    -v "/var/run/docker.sock:/var/run/docker.sock" \
    -e "GH_SHOW_DIFFS=${GH_SHOW_DIFFS:-false}" \
    -e "GH_FIX=${GH_FIX:-false}" \
    -w /data \
    githooks:test-rules \
    nix --accept-flake-config develop ".#default" --command tests/exec-rules.sh ||
    {
        echo "! Check rules had failures: exit code: $?"
        exit 1
    }
