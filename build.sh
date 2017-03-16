#!/bin/bash

TARGET_FILENAME="az_healthcheck"

if [ ! -d target ]; then
  mkdir -p target
fi

for GOOS in darwin linux windows; do
  for GOARCH in 386 amd64; do
    echo "Building $GOOS-$GOARCH"
    export GOOS=$GOOS
    export GOARCH=$GOARCH
    go build -o target/${TARGET_FILENAME}-$GOOS-$GOARCH ${TARGET_FILENAME}.go
  done
done

