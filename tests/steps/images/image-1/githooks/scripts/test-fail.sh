#!/usr/bin/env bash

[ "$GITHOOKS_CONTAINER_RUN" = "true" ] || {
    echo "Should be inside container!" >&2
    exit 222
}

echo "Executing test script."
exit 123
