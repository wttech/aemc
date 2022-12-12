#!/usr/bin/env sh

. aem/down.sh

step "destroying instances"
aem instance destroy

