#!/usr/bin/env sh

UNIT=${1:-author}

docker run -it --rm \
  --platform linux/x86_64 \
  -v "$(pwd)/src/aem/default:/opt/aemc/aem/default" \
  "acme/aem/${UNIT}"
