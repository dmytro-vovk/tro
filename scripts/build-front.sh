#!/bin/bash

set -eu

docker run \
  --rm \
  --volume $(PWD):/src \
  --workdir /src \
  node:alpine3.14 \
  npm install

docker run \
  --rm \
  --volume $(PWD):/src \
  --workdir /src \
  node:alpine3.14 \
  npm run build

rm -rf ./internal/webserver/handlers/home/css

cp -rf ./frontend/styles ./internal/webserver/handlers/home/css

gzip -c ./frontend/index.html > ./internal/webserver/handlers/home/index.html.gz

gzip -f ./internal/webserver/handlers/home/index.js

gzip -f ./internal/webserver/handlers/home/index.js.map
