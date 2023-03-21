#!/usr/bin/env sh

UNIT=${1:-base}

docker build --progress=plain --platform linux/x86_64 -t "acme/aem/${UNIT}" -f "${UNIT}.Dockerfile" .
