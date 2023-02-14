#!/usr/bin/env sh

rm -fr etc.src && cp -r ../src etc.src && docker build -t aemc/dispatcher-publish:latest .
