#!/bin/bash

docker run \
  --rm \
  --volume "$(PWD)":/src \
  --workdir /src \
  golang:1.17 \
  go build -ldflags "-s -w" -o app cmd/main.go
