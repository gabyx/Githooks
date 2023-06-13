#!/usr/bin/env bash

echo "Arguments given: '$1', '$2'"

for file in $STAGED_FILES; do
    echo "Formatting file '$file'..."
    echo "Formatted by containerized hook" >>"$file"
done
