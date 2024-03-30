#!/usr/bin/env bash

set -e
set -u

function clean_up() {
    # shellcheck disable=SC2317
    docker rmi "githooks:unittests" &>/dev/null || true
}

trap clean_up EXIT

cat <<EOF | docker build --force-rm -t githooks:unittests -
FROM golang:1.20-alpine
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

RUN chown -R "test-user:test-user" "/home/test-user/.config/containers" && \
    chown -R "test-user:test-user" "/home/test-user/.local/share/containers"

USER "test-user"

# CVE https://github.blog/2022-10-18-git-security-vulnerabilities-announced/#cve-2022-39253
RUN git config --global protocol.file.allow always
# Git refuses to do stuff in this mounted directory.
RUN git config --global safe.directory /githooks

ENV DOCKER_RUNNING=true
EOF

if ! docker run --privileged --rm -it \
    -v "$(pwd)":/githooks \
    -w /githooks githooks:unittests \
    tests/exec-unittests.sh ".*Podman.*" "test_podman"; then

    echo "! Check rules had failures."
    exit 1
fi

exit 0
