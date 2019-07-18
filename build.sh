#!/bin/bash

TARGET_FILENAME="azhealthcheck"

if [ ! -d target ]; then
  mkdir -p target
else
  rm -f target/* 1>/dev/null 2>&1
fi


TARGET_OPERATING_SYSTEMS="darwin linux windows"
TARGET_PLATFORMS="amd64" # i386

for GOOS in $TARGET_OPERATING_SYSTEMS; do
  for GOARCH in $TARGET_PLATFORMS; do
    echo "Building $GOOS-$GOARCH"
    export GOOS=$GOOS
    export GOARCH=$GOARCH
    FINAL_FILENAME="target/${TARGET_FILENAME}-$GOOS-$GOARCH"
    if [ $GOOS == "windows" ]; then
      FINAL_FILENAME="${FINAL_FILENAME}.exe"
    fi
    go build -o $FINAL_FILENAME .
  done
done

