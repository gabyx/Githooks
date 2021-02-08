#!/bin/sh
# Base Git hook template from https://github.com/gabyx/githooks
#
# It allows you to have a .githooks folder per-project that contains
# its hooks to execute on various Git triggers.

# Read the runner script from the local/global or system config
GITHOOKS_RUNNER=$(git config githooks.runner)

if [ ! -x "$GITHOOKS_RUNNER" ]; then
    echo "! Githooks runner points to a non existing location" >&2
    echo "   \`$GITHOOKS_RUNNER\`" >&2
    echo " or it is not executable!" >&2
    echo " Please run the Githooks install script again to fix it." >&2
    exit 1
fi

[ -z "$GH_COVERAGE_DIR" ] && {
    echo "! Env variables 'GH_COVERAGE_DIR' not set"
    exit 1
}

COV_DATA="$GH_COVERAGE_DIR/runner.yaml"
COUNTER=$(head -1 "$COV_DATA" | sed -E 's@counter: ([0-9]+)@\1@')
[ -z "$COUNTER" ] && COUNTER="0"
COV_FILE="$GH_COVERAGE_DIR/runner-$COUNTER.cov"
[ -f "$COV_FILE" ] && {
    echo "! Coverage file '$COV_FILE' already existing."
    exit 1
}
echo "Writting to '$COV_FILE'"
exec "$GITHOOKS_RUNNER" -test.coverprofile "$COV_FILE" githooksCoverage "$0" "$@"
