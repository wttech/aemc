#!/usr/bin/env sh

UNIT=${1:-author}

docker build --progress=plain --no-cache -t "acme/aem/${UNIT}" -f "${UNIT}.Dockerfile" .
