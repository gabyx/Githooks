#!/bin/sh
# Test:
#   Cli tool: manage local shared hook repositories

git config --global githooks.testingTreatFileProtocolAsRemote "true"

if ! /var/lib/githooks/githooks/bin/installer --stdin; then
    echo "! Failed to execute the install script"
    exit 1
fi

mkdir -p /tmp/shared/first-shared.git/.githooks/pre-commit &&
    mkdir -p /tmp/shared/second-shared.git/.githooks/pre-commit &&
    mkdir -p /tmp/shared/third-shared.git/.githooks/pre-commit &&
    echo 'echo "Hello"' >/tmp/shared/first-shared.git/.githooks/pre-commit/sample-one &&
    echo 'echo "Hello"' >/tmp/shared/second-shared.git/.githooks/pre-commit/sample-two &&
    echo 'echo "Hello"' >/tmp/shared/third-shared.git/.githooks/pre-commit/sample-three &&
    (cd /tmp/shared/first-shared.git && git init && git add . && git commit -m 'Testing') &&
    (cd /tmp/shared/second-shared.git && git init && git add . && git commit -m 'Testing') &&
    (cd /tmp/shared/third-shared.git && git init && git add . && git commit -m 'Testing') ||
    exit 1

mkdir -p /tmp/test083 && cd /tmp/test083 && git init || exit 1

testShared() {
    git hooks shared add --shared file:///tmp/shared/first-shared.git &&
        git hooks shared list | grep "first-shared" | grep "pending" &&
        git hooks shared pull &&
        git hooks shared list | grep "first-shared" | grep "active" &&
        git hooks shared add --shared file:///tmp/shared/second-shared.git &&
        git hooks shared add file:///tmp/shared/third-shared.git &&
        git hooks shared list --shared | grep "second-shared" | grep "pending" &&
        git hooks shared list --all | grep "third-shared" | grep "pending" &&
        (cd ~/.githooks/shared/*shared-first-shared-git* &&
            git remote rm origin &&
            git remote add origin /some/other/url.git) &&
        git hooks shared list | grep "first-shared" | grep "invalid" &&
        git hooks shared remove --shared file:///tmp/shared/first-shared.git &&
        ! git hooks shared list | grep "first-shared" &&
        git hooks shared remove --shared file:///tmp/shared/second-shared.git &&
        git hooks shared remove file:///tmp/shared/third-shared.git &&
        [ ! -f "$(pwd)/.githooks/.shared" ] ||
        exit "$1"
}

testShared 2

git hooks shared clear --all &&
    git hooks shared purge ||
    exit 8

testShared 9
