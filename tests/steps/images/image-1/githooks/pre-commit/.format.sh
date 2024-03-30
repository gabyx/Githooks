#!/bin/bash
set -u
set -e

echo "Arguments given:" "$@"

if [ "${1:-}" != "--banana" ]; then
    echo "! First argument is not --banana."
    exit 1
fi

if [ "${MONKEY:-}" != "gaga" ]; then
    echo "! Env. variable MONKEY is not defined or wrong value."
    exit 1
fi

function do_stuff() {
    local file="$1"
    if [ ! -f "$file" ]; then
        echo "File '$file' does not exist!"
        exit 1
    fi
    echo "Formatting file '$file'..."
    echo "Formatted by containerized hook" >>"$file"
}

# Make sure we really do the right thing for the test.
if [ ! -f .githooks-test-export-staged-files ]; then
    [ -z "${STAGED_FILES_FILE:-}" ] || {
        echo "STAGED_FILES_FILE must not be defined."
        exit 1
    }

    for file in $STAGED_FILES; do
        do_stuff "$file"
    done
else
    [ -z "${STAGED_FILES:-}" ] || {
        echo "STAGED_FILES must not be defined."
        exit 1
    }

    while read -rd $'\0' file; do
        do_stuff "$file"
    done <"$STAGED_FILES_FILE"
fi
