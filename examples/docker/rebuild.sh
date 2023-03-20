#!/usr/bin/env sh

UNIT=${1:-base}

docker build --progress=plain --no-cache -t "acme/aem/${UNIT}" -f "${UNIT}.Dockerfile" .
