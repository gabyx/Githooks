#!/usr/bin/env bash
# shellcheck disable=SC1091

set -e
set -u

rootDir=$(git rev-parse --show-toplevel)
. "$rootDir/tests/general.sh"

imageType="$1"
shift

if echo "$imageType" | grep -q "\-user"; then
    os_user="test"
else
    os_user="root"
fi

# shellcheck disable=SC2317
function clean_up() {
    docker rmi "githooks:$imageType" &>/dev/null || true
    docker rmi "githooks:$imageType-base" &>/dev/null || true
    docker volume rm gh-test-tmp &>/dev/null || true
}

trap clean_up EXIT

function buildImage() {
    local dockerGroupId="$1"

    # Build container to only copy to volumes.
    cat <<EOF | docker build \
        --force-rm -t "githooks:volumecopy" -f - . || exit 1
    FROM scratch
    CMD you-should-not-run-this-container
EOF

    # Build the Githooks test container.
    cat <<EOF | docker build \
        --build-arg "DOCKER_GROUP_ID=$dockerGroupId" \
        --force-rm -t "githooks:$imageType" -f - "$rootDir" || exit 1

FROM githooks:$imageType-base
ARG DOCKER_GROUP_ID

ENV DOCKER_RUNNING=true
ENV GH_TESTS="/var/lib/githooks-tests"
ENV GH_TEST_TMP="/tmp/githooks"
ENV GH_TEST_REPO="/var/lib/githooks"
ENV GH_TEST_BIN="/var/lib/githooks/githooks/bin"
ENV GH_TEST_GIT_CORE="/usr/share/git-core"

${ADDITIONAL_PRE_INSTALL_STEPS:-}

# Add sources
COPY --chown=$os_user:$os_user githooks "\$GH_TEST_REPO/githooks"
ADD .githooks/README.md "\$GH_TEST_REPO/.githooks/README.md"
ADD examples "\$GH_TEST_REPO/examples"

# Setup Githooks
ADD tests/setup-githooks.sh "\$GH_TESTS/"
RUN bash "\$GH_TESTS/setup-githooks.sh"

# Add all tests
ADD tests "\$GH_TESTS"

# Modify install arguments.
RUN if [ -n "\$EXTRA_INSTALL_ARGS" ]; then \\
        sed -i -E 's|(.*)/cli\" installer|\1/cli" installer \$EXTRA_INSTALL_ARGS|g' "\$GH_TESTS"/steps/step-* ; \\
    fi

# Always don't delete LFS Hooks (for testing, default is unset, but cumbersome for tests)
RUN git config --global githooks.deleteDetectedLFSHooks "n"

# Git-Core folder must be existing.
RUN [ -d "\$GH_TEST_GIT_CORE/templates/hooks" ]

${ADDITIONAL_INSTALL_STEPS:-}

RUN echo "Git version: \$(git --version)"
WORKDIR \$GH_TESTS
EOF

}

# Only works on linux (macOS does not need it)
dockerGroupId=$(getent group docker 2>/dev/null | cut -d: -f3) || true
echo "Docker group id: $dockerGroupId"
buildImage "$dockerGroupId"

# Create a volume where all test setup and repositories go in.
# Is mounted to `/tmp`.
deleteContainerVolume gh-test-tmp
docker volume create gh-test-tmp

# Privileged --privileged is needed if you want
# launch nested containers when not sharing the docker socket.
# Both are dangerous and should be handled with care.
docker run \
    --privileged \
    --rm \
    -a stdout -a stderr \
    -v "gh-test-tmp:/tmp" \
    -v "/var/run/docker.sock:/var/run/docker.sock" \
    "githooks:$imageType" \
    bash ./exec-steps.sh "$@"
