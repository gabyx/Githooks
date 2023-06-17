#!/bin/bash
set -u
set -e

echo "Arguments given:" "$@"

if [ "${1:-}" != "--banana" ]; then
    echo "! First argument is not --banana"
    exit 1
fi

for file in $STAGED_FILES; do
    if [ ! -f "$file" ]; then
        echo "File '$file' does not exist!"
        exit 1
    fi
    echo "Formatting file '$file'..."
    echo "Formatted by containerized hook" >>"$file"
done
