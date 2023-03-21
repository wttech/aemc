#!/usr/bin/env sh

UNIT=${1:-author}

docker build --platform linux/x86_64 -t "acme/aem/${UNIT}" -f "${UNIT}.Dockerfile" .
