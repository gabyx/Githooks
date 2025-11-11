#!/usr/bin/env bash
# shellcheck disable=SC1091

set -e
set -u

ROOT_DIR=$(git rev-parse --show-toplevel)
. "$ROOT_DIR/tests/general.sh"

cd "$ROOT_DIR"

cat <<'EOF' | docker build --force-rm -t githooks:windows-lfs -f - .
FROM mcr.microsoft.com/dotnet/framework/runtime:4.8-windowsservercore-ltsc2022


# $ProgressPreference: https://github.com/PowerShell/PowerShell/issues/2138#issuecomment-251261324
SHELL ["powershell", "-Command", "$ErrorActionPreference = 'Stop';"]

RUN iex ((New-Object System.Net.WebClient).DownloadString('https://chocolatey.org/install.ps1'))
RUN choco install -y git
RUN choco install -y jq
RUN choco install -y curl

# CVE https://github.blog/2022-10-18-git-security-vulnerabilities-announced/#cve-2022-39253
RUN git config --system protocol.file.allow always

# ideally, this would be C:\go to match Linux a bit closer, but C:\go is the recommended install path for Go itself on Windows
ENV GOPATH C:\\gopath

# PATH isn't actually set in the Docker image, so we have to set it from within the container
RUN $newPath = ('{0}\bin;C:\go\bin;{1}' -f $env:GOPATH, $env:PATH); \
    Write-Host ('Updating PATH: {0}' -f $newPath); \
    [Environment]::SetEnvironmentVariable('PATH', $newPath, [EnvironmentVariableTarget]::Machine);
# doing this first to share cache across versions more aggressively

# Check hash below for download.
ENV GOLANG_VERSION 1.24.10

RUN $url = ('https://go.dev/dl/go{0}.windows-amd64.zip' -f $env:GOLANG_VERSION); \
    Write-Host ('Downloading {0} ...' -f $url); \
    $ProgressPreference = 'SilentlyContinue'; Invoke-WebRequest -Uri $url -OutFile 'go.zip'; \
    \
    $sha256 = '78d4ea375b9f729c4883e0c1a92d63c73f1bcf0edaafdb0932295472f72acbce'; \
    Write-Host ('Verifying sha256 ({0}) ...' -f $sha256); \
    $sha256Ex = (Get-FileHash go.zip -Algorithm sha256).Hash; \
    if ($sha256Ex -ne $sha256) { \
        Write-Host ('FAILED! Got sha: {0}' -f $sha256Ex); \
        exit 1; \
    }; \
    \
    Write-Host 'Expanding ...'; \
    $ProgressPreference = 'SilentlyContinue'; Expand-Archive go.zip -DestinationPath C:\; \
    \
    Write-Host 'Removing ...'; \
    Remove-Item go.zip -Force; \
    \
    Write-Host 'Verifying install ("go version") ...'; \
    go version; \
    \
    Write-Host 'Complete.';

ENV DOCKER_RUNNING=true
ENV GH_SCRIPTS="c:/githooks-tests/scripts"
ENV GH_TESTS="c:/githooks-tests/tests"
ENV GH_TEST_TMP="c:/githooks-tests/tmp"
ENV GH_TEST_REPO="c:/githooks-tests/githooks"
ENV GH_TEST_BIN="c:/githooks-tests/githooks/githooks/bin"
ENV GH_TEST_GIT_CORE="c:/Program Files/Git/mingw64/share/git-core"
ENV GH_ON_WINDOWS="true"

# Add sources
COPY githooks "$GH_TEST_REPO/githooks"
ADD .githooks/README.md "$GH_TEST_REPO/.githooks/README.md"
ADD examples "$GH_TEST_REPO/examples"

ADD tests/setup-githooks.sh "$GH_TESTS/"
RUN & "'C:/Program Files/Git/bin/bash.exe'" "C:/githooks-tests/tests/setup-githooks.sh"

ADD tests "$GH_TESTS"
ADD scripts "$GH_SCRIPTS"

WORKDIR C:/githooks-tests/tests

EOF

docker run --rm \
    -a stdout \
    -a stderr "githooks:windows-lfs" \
    "C:/Program Files/Git/bin/sh.exe" ./exec-steps.sh --skip-docker-check "$@"

RESULT=$?

docker rmi "githooks:windows-lfs"
exit $RESULT
