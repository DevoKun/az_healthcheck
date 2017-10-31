#!/bin/bash

TARGET_FILENAME="azhealthcheck"

if [ ! -d target ]; then
  mkdir -p target
else
  rm -f target/* 1>/dev/null 2>&1
fi


TARGET_OPERATING_SYSTEMS="darwin linux" # windows
TARGET_PLATFORMS="amd64" # i386

for GOOS in $TARGET_OPERATING_SYSTEMS; do
  for GOARCH in $TARGET_PLATFORMS; do
    echo "Building $GOOS-$GOARCH"
    export GOOS=$GOOS
    export GOARCH=$GOARCH
    go build -o target/${TARGET_FILENAME}-$GOOS-$GOARCH ${TARGET_FILENAME}.go
  done
done

