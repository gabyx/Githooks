#!/bin/sh
# Test:
#   Direct template execution: fail on shared hooks

git config --global githooks.testingTreatFileProtocolAsRemote "true"

mkdir -p ~/.githooks/release && cp /var/lib/githooks/*.sh ~/.githooks/release || exit 1
mkdir -p /tmp/shared/hooks-018.git/pre-commit &&
    echo 'exit 1' >/tmp/shared/hooks-018.git/pre-commit/fail &&
    cd /tmp/shared/hooks-018.git &&
    git init &&
    git add . &&
    git commit -m 'Initial commit' ||
    exit 1

mkdir -p /tmp/test18 && cd /tmp/test18 || exit 1
git init || exit 1

mkdir -p .githooks &&
    echo 'file:///tmp/shared/hooks-018.git' >.githooks/.shared &&
    ~/.githooks/release/cli.sh shared update ||
    exit 1

~/.githooks/release/base-template.sh "$(pwd)"/.git/hooks/pre-commit

if [ $? -ne 1 ]; then
    echo "! Expected to fail on shared hook execution"
    exit 1
fi
