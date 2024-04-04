#!/bin/sh
# Base Git hook template from https://github.com/gabyx/githooks
#
# It allows you to have a .githooks folder per-project that contains
# its hooks to execute on various Git triggers.

# Read the runner script from the local/global or system config
GITHOOKS_RUNNER=$(command -v "githooks-runner" 2>/dev/null || git config githooks.runner)

# shellcheck disable=SC2181
if [ ! -x "$GITHOOKS_RUNNER" ]; then
    echo "! Either 'githooks-runner' must be in your path or" >&2
    echo "  Git config value in 'githooks.runner' must point to an " >&2
    echo "  executable. The value:" >&2
    echo "   '$GITHOOKS_RUNNER" >&2
    echo "  is not existing or is not executable!" >&2
    echo "  Please run the Githooks install script again to fix it." >&2
    exit 1
fi

[ -z "$GH_COVERAGE_DIR" ] && {
    echo "! Env variables 'GH_COVERAGE_DIR' not set" >&2
    exit 1
}

COV_DATA="$GH_COVERAGE_DIR/githooks-runner.yaml"
COUNTER=$(head -1 "$COV_DATA" 2>&1 | sed -E 's@counter: ([0-9]+)@\1@')
[ -z "$COUNTER" ] && COUNTER="0"
COV_FILE="$GH_COVERAGE_DIR/runner-$COUNTER.cov"
[ -f "$COV_FILE" ] && {
    echo "! Coverage file '$COV_FILE' already existing." >&2
    exit 1
}
echo "Writing to '$COV_FILE'" >&2
exec "$GITHOOKS_RUNNER" -test.coverprofile "$COV_FILE" githooksCoverage "$0" "$@"
