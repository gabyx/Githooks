#!/bin/sh
# Base Git hook template from https://github.com/rycus86/githooks
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

exec "$GITHOOKS_RUNNER" "$0" "$@"
