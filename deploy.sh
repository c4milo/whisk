#!/usr/bin/env bash
# This script is called from https://jenkins.tinyspeck.com/job/whisk
set -o noclobber  # Avoid overlay files (echo "hi" > foo)
set -o pipefail   # Unveils hidden failures
set -o nounset    # Exposes unset variables

CURRENT_PATH=$(pwd)
SERVICE_PATH="whisk/cmd/whisk"
OUTPUT_FILENAME=$(basename "$SERVICE_PATH")
OUTPUT_PATH="$CURRENT_PATH/$OUTPUT_FILENAME"

LATEST_VERSION=$(aws s3api list-objects-v2 --bucket=slack-goslackgo --prefix=versions_by_date/goslackgo_master --query='Contents[-1].Key' --output=text | xargs basename)
S3_PATH="s3://slack-goslackgo/artifacts/$LATEST_VERSION/linux_amd64/$SERVICE_PATH/$OUTPUT_FILENAME.bin"

printf "retrieving %s...\n", "$S3_PATH"
aws s3 cp "$S3_PATH" "$OUTPUT_PATH"
chmod +x "$OUTPUT_PATH"
printf "binary downloaded and saved to %s" "$OUTPUT_PATH\n"
