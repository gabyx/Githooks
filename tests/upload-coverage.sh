#!/bin/bash
TEST_DIR=$(cd "$(dirname "$0")" && pwd)

[ -f "$TEST_DIR/cover/all.cov" ] || {
    echo "! No coverage file existing" >&2
    exit 1
}

# shellcheck disable=SC2015
cd "githooks" &&
    scripts/build.sh && # Generate all files again such that we can upload the coverage
    goveralls -coverprofile="$TEST_DIR/cover/all.cov" -service=travis-ci || {
    echo "! Goveralls failed." >&2
    exit 1
}

exit 0
