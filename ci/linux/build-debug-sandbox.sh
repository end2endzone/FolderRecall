#!/usr/bin/env bash
set -e

# Get the projet's root directory
PROJECTROOT=$(cd "$(dirname "$0")/../.." && pwd)

cd $PROJECTROOT

rm -rf .debug-sandbox
mkdir -p .debug-sandbox
cp -r testdata .debug-sandbox

echo "Sandbox ready!"
