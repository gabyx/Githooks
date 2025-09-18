#!/usr/bin/env bash
# shellcheck disable=SC1091

set -e
set -u

ROOT_DIR=$(git rev-parse --show-toplevel)
. "$ROOT_DIR/tests/general.sh"

cd "$ROOT_DIR"

function clean_up() {
    # shellcheck disable=SC2317
    run_docker rmi "githooks:unittests" &>/dev/null || true
}

trap clean_up EXIT

cd "$ROOT_DIR"

cat <<EOF | run_docker build --force-rm -t githooks:unittests -f - "$ROOT_DIR"
FROM golang:1.22-alpine
RUN apk update && apk add git git-lfs
RUN apk add bash jq curl

RUN adduser "test-user" \
    -D \
    -u "1000" -g "1000" \
    -h "/home/test-user"

RUN apk add crun shadow openrc fuse-overlayfs shadow slirp4netns
RUN apk add podman

RUN echo "test-user:100000:65536" > /etc/subuid && \
    echo "test-user:100000:65536" > /etc/subgid

RUN temp=\$(mktemp) && \
    sed -E 's/rc_cgroup_mode=.*/rc_cgroup_mode="unified"/g' /etc/rc.conf >"\$temp" && \
    mv "\$temp" /etc/rc.conf

RUN rc-service cgroups start || true
RUN rc-update add cgroups

RUN mkdir -p "/home/test-user/.config/containers" && \
    mkdir -p "/home/test-user/.local/share/containers"

RUN (echo '[containers]' && \
    echo 'netns="host"' && \
    echo 'userns="host"' && \
    echo 'ipcns="host"' && \
    echo 'utsns="host"' && \
    echo 'cgroupns="host"' && \
    echo 'cgroups="disabled"' && \
    echo 'log_driver = "k8s-file"' && \
    echo '[engine]' && \
    echo 'cgroup_manager = "cgroupfs"' && \
    echo 'events_logger="file"' && \
    echo 'runtime="crun"') >/etc/containers/containers.conf

RUN (echo "[containers]" && \
    echo "volumes = [" && \
    echo "  \"/proc:/proc\"," && \
    echo "]" && \
    echo "default_sysctls = []") >"/home/test-user/.config/containers/containers.conf"

RUN (echo "[storage]" && \
    echo "driver = \"overlay\"" && \
    echo "[storage.options.overlay]" && \
    echo "mount_program = \"\$(which fuse-overlayfs)\"") >"/home/test-user/.config/containers/storage.conf"

RUN chown -R "test-user:test-user" "/home/test-user/.config" && \
    chown -R "test-user:test-user" "/home/test-user/.local/share/containers"

USER "test-user"

# Git refuses to do stuff in directories.
# also CVE https://github.blog/2022-10-18-git-security-vulnerabilities-announced/#cve-2022-39253
RUN git config --global safe.directory '*' && \
    git config --global user.email test@githooks.com && \
    git config --global user.name Githooks CI && \
    git config --global protocol.file.allow always

RUN mkdir -p ~/repo/
COPY --chown=1000:1000 ./githooks /home/test-user/repo/githooks
COPY --chown=1000:1000 ./tests /home/test-user/repo/tests
RUN cd ~/repo && git init . && git add . && git commit -a -m init && git tag v1.0.0
WORKDIR /home/test-user/repo

ENV DOCKER_RUNNING=true
EOF

if ! run_docker run --privileged --rm -it \
    githooks:unittests \
    tests/exec-unittests.sh ".*Podman.*" "test_podman"; then
    echo "! Check rules had failures."
    exit 1
fi

exit 0
