#!/usr/bin/env sh

UNIT=${1:-base}

docker build --no-cache -t "acme/aem/${UNIT}" -f "${UNIT}.Dockerfile" .
