#!/usr/bin/env bash
set -euo pipefail

# Assert required variables are defined
REQUIRED_VARS=("GH_TOKEN" "GIST_ID" "FILE_PATH" "GIST_FILENAME")

for var in "${REQUIRED_VARS[@]}"; do
  if [ -z "${!var:-}" ]; then
    echo "Error: Mandatory environment variable '$var' is not set." >&2
    exit 1
  fi
done

# Assert local file actually exists before trying to upload
if [ ! -f "$FILE_PATH" ]; then
  echo "Error: Local file standard path '$FILE_PATH' does not exist." >&2
  exit 1
fi

# Upload (create new file or override existing)
echo "Checking Gist ($GIST_ID) for file: $GIST_FILENAME..."
if gh gist view "$GIST_ID" --files | grep -qx "$GIST_FILENAME"; then
  echo "File '$GIST_FILENAME' exists in Gist. Overwriting..."
  gh gist edit "$GIST_ID" "$FILE_PATH" --filename "$GIST_FILENAME"
else
  echo "File '$GIST_FILENAME' does not exist in Gist. Adding..."
  gh gist edit "$GIST_ID" "$FILE_PATH" --add "$GIST_FILENAME"
fi

echo "Successfully updated Gist!"