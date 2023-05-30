#!/usr/bin/env bash
ROOT_DIR=$(git rev-parse --show-toplevel)

IMAGE_TYPE="$1"
shift

if echo "$IMAGE_TYPE" | grep -q "\-user"; then
    OS_USER="test"
else
    OS_USER="root"
fi

# Only works on linux (macOS does not need it)
dockerGroupId=$(getent group docker 2>/dev/null | cut -d: -f3) || true
echo "Docker group id: $dockerGroupId"

cat <<EOF | docker build \
    --build-arg "DOCKER_GROUP_ID=$dockerGroupId" \
    --force-rm -t "githooks:$IMAGE_TYPE" -f - "$ROOT_DIR" || exit 1

FROM githooks:$IMAGE_TYPE-base
ARG DOCKER_GROUP_ID

ENV DOCKER_RUNNING=true
ENV GH_TESTS="/var/lib/githooks-tests"
ENV GH_TEST_TMP="/tmp"
ENV GH_TEST_REPO="/var/lib/githooks"
ENV GH_TEST_BIN="/var/lib/githooks/githooks/bin"
ENV GH_TEST_GIT_CORE="/usr/share/git-core"

${ADDITIONAL_PRE_INSTALL_STEPS:-}

# Add sources
COPY --chown=$OS_USER:$OS_USER githooks "\$GH_TEST_REPO/githooks"
ADD .githooks/README.md "\$GH_TEST_REPO/.githooks/README.md"
ADD examples "\$GH_TEST_REPO/examples"

ADD tests/setup-githooks.sh "\$GH_TESTS/"
RUN bash "\$GH_TESTS/setup-githooks.sh"

ADD tests "\$GH_TESTS"

RUN if [ -n "\$EXTRA_INSTALL_ARGS" ]; then \\
        sed -i -E 's|(.*)/cli\" installer|\1/cli" installer \$EXTRA_INSTALL_ARGS|g' "\$GH_TESTS"/step-* ; \\
    fi

# Always don't delete LFS Hooks (for testing, default is unset, but cumbersome for tests)
RUN git config --global githooks.deleteDetectedLFSHooks "n"

${ADDITIONAL_INSTALL_STEPS:-}

RUN echo "Git version: \$(git --version)"
WORKDIR \$GH_TESTS
EOF

runArgs=()
if [ "${CI:-}" != "true" ]; then
    runArgs=("-it")
fi

docker run --rm \
    -a stdout -a stderr \
    "${runArgs[@]}" \
    -v "/var/run/docker.sock:/var/run/docker.sock" \
    "githooks:$IMAGE_TYPE" \
    ./exec-steps.sh "$@"

RESULT=$?

docker rmi "githooks:$IMAGE_TYPE"
docker rmi "githooks:$IMAGE_TYPE-base"
exit $RESULT
