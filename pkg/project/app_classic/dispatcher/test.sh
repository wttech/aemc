#!/usr/bin/env sh

# Use this script to test built image by running commands in container shell

docker run --rm -it --entrypoint bash acme/aem-ams/dispatcher-publish:latest
