#!/usr/bin/env sh

UNIT=${1:-base}

docker build -t "acme/aem/${UNIT}" -f "${UNIT}.Dockerfile" .
