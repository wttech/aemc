#!/usr/bin/env sh

rm -fr src.origin && cp -r ../src src.origin && docker build -t aemc/dispatcher-publish:latest .
