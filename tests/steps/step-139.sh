#!/usr/bin/env bash
# Test:
#   Run a simple install and install,check env. vars in `installDir`.
#   https://github.com/gabyx/Githooks/issues/142

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

acceptAllTrustPrompts || exit 1

# run the default install
"$GH_TEST_BIN/cli" installer || exit 1

installDir=$(git config githooks.installDir)
# shellcheck disable=SC2088,SC2016
if ! echo "$installDir" | grep '\$HOME'; then
    echo "! Expected ~/ to be part of install dir: $installDir"
    exit 1
fi

# Make some whitespace changes to the global gitconfig
# to test that it does not get updated.
sed -i -E "s/githooks\.installDir/    githooks\.installDir/g" ~/.gitconfig || exit 1
mkdir -p "$GH_TEST_TMP/test139" && cp ~/.gitconfig "$GH_TEST_TMP/test139/"

# Set server to 9.9.1 to trigger update.
if ! git -C "$GH_TEST_REPO" reset --hard v9.9.1 >/dev/null; then
    echo "! Could not reset server to trigger update."
    exit 1
fi

# Update to version 9.9.1
echo "Update to version 9.9.1"
CURRENT="$(git -C ~/.githooks/release rev-parse HEAD)"
if ! "$GH_INSTALL_BIN_DIR/cli" update --yes; then
    echo "! Failed to run the update"
fi
AFTER="$(git -C ~/.githooks/release rev-parse HEAD)"

if [ "$CURRENT" = "$AFTER" ] ||
    [ "$(git -C "$GH_TEST_REPO" rev-parse v9.9.1)" != "$AFTER" ]; then
    echo "! Release clone was not updated, but it should have!"
    exit 1
fi

if ! git diff --exit-code \
    ~/.gitconfig "$GH_TEST_TMP/test139/.gitconfig"; then

    echo "! Update suddenly changes the Git config, it should not have changed: Output:"
    git diff ~/.gitconfig "$GH_TEST_TMP/test139/.gitconfig"

    exit 1
fi

# Install again with prefix and check if the raw entered install directory is
# maintained.
"$GH_TEST_BIN/cli" installer --prefix "\$GH_TEST_TMP/test139" || exit 1

installDir=$(git config githooks.installDir)
if ! echo "$installDir" | grep "\$GH_TEST_TMP"; then
    echo "! Expected \$GH_TEST_TMP to be part of install dir: $installDir"
    cat ~/.gitconfig
    exit 1
fi
