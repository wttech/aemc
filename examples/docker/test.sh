#!/usr/bin/env sh

UNIT=${1:-base}

docker run --platform linux/x86_64 -it "acme/aem/${UNIT}"
