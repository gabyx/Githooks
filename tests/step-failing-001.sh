#!/bin/sh

# run the install and try to install hooks into a non-existing directory
rm -rf /does/not/exist

echo 'y
/does/not/exist
' | sh /var/lib/githooks/install.sh

if [ $? -eq 0 ]; then
    echo "! Expected to fail"
    exit 1
fi

