#!/bin/bash

[ -d "$GH_COVERAGE_DIR" ] || {
    echo "! No coverage dir existing" >&2
    exit 1
}

if [ -n "$GH_COVERAGE_DIR" ]; then
    # shellcheck disable=SC2015
    gocovmerge "$GH_COVERAGE_DIR"/*.cov >"$GH_COVERAGE_DIR/all.cov" || {
        echo "! Cov merge failed." >&2
        exit 1
    }
    echo "Coverage created."
fi

# shellcheck disable=SC2015
cd "githooks" &&
    scripts/build.sh && # Generate all files again such that we can upload the coverage
    goveralls -coverprofile="$GH_COVERAGE_DIR/all.cov" -service=travis-ci || {
    echo "! Goveralls failed." >&2
    exit 1
}

exit 0
