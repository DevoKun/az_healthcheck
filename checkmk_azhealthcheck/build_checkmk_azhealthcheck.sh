#!/bin/bash

export GOOS=darwin
export GOARCH=amd64

go build -o checkmk_azhealthcheck checkmk_azhealthcheck.go
