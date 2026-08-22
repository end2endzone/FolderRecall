#!/usr/bin/env bash
set -e

# Get the projet's root directory
PROJECTROOT=$(cd "$(dirname "$0")/../.." && pwd)

cd $PROJECTROOT

# GOOS
if [[ -n "${GOOS}" ]]; then
    echo "GOOS is set to: $GOOS"
else
    GOOS=$(go env GOHOSTOS)
    echo "GOOS is not set and is forced to: $GOOS"
fi

# GOARCH
if [[ -n "${GOARCH}" ]]; then
    echo "GOARCH is set to: $GOARCH"
else
    GOARCH=$(go env GOHOSTARCH)
    echo "GOARCH is not set and is forced to: $GOARCH"
fi

# Define the target binary file path based on the environment
TARGET="$PROJECTROOT/bin/fldrecall"
if [[ "$CI" == "true" ]]; then
    TARGET="$PROJECTROOT/bin/fldrecall-$GOOS-$GOARCH"
    echo "Building on CI/CD server. Changing the target file name to '$TARGET'"
fi

# Show compiled version
$TARGET --version --verbose
