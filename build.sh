#!/usr/bin/env bash

set -o noclobber  # Avoid overlay files (echo "hi" > foo)
set -o errexit    # Used to exit upon error, avoiding cascading errors
set -o pipefail   # Unveils hidden failures
set -o nounset    # Exposes unset variables

export GOARCH='amd64'
export CGO_ENABLED=0
for GOOS in linux darwin; do
  export GOOS=$GOOS
  go build \
    -a -tags netgo -ldflags '-w -extldflags "-static"' \
    -o build/slack-whisk-$GOOS-$GOARCH.bin \
    cmd/whisk/main.go
done
